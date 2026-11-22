// Package seeds finds bootstrap peers for the LIBR DHT. Sources, in order: explicit
// configuration, the list shipped in this binary, peers cached from earlier runs, and
// seed servers (librserver). Seed servers only hand out addresses; anything learned
// from them is checked by the peer ID it is dialed under (Noise), so a lying server
// can waste a dial but cannot impersonate a node.
package seeds

import (
	"bytes"
	"context"
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/libr-forum/Libr/core/libr/proto"
)

//go:embed seeds.json
var shipped []byte

const (
	maxAddrs    = 8 // per registration, as librserver enforces
	maxCached   = 32
	maxReply    = 64 << 10
	httpTimeout = 10 * time.Second
)

var client = &http.Client{Timeout: httpTimeout}

// Shipped returns the seeds built into this binary ("/.../p2p/<id>" multiaddrs).
func Shipped() []peer.AddrInfo {
	var list []string
	_ = json.Unmarshal(shipped, &list)
	return Parse(list)
}

// Parse turns "/.../p2p/<id>" strings into AddrInfos, merging addresses of the same
// peer and skipping anything malformed.
func Parse(list []string) []peer.AddrInfo {
	var out []peer.AddrInfo
	for _, s := range list {
		ai, err := peer.AddrInfoFromString(strings.TrimSpace(s))
		if err != nil {
			continue
		}
		out = Merge(out, []peer.AddrInfo{*ai})
	}
	return out
}

// Merge appends b to a, combining addresses of peers already in a.
func Merge(a, b []peer.AddrInfo) []peer.AddrInfo {
	idx := map[peer.ID]int{}
	for i, ai := range a {
		idx[ai.ID] = i
	}
	for _, ai := range b {
		if i, ok := idx[ai.ID]; ok {
			a[i].Addrs = append(a[i].Addrs, ai.Addrs...)
			continue
		}
		idx[ai.ID] = len(a)
		a = append(a, ai)
	}
	return a
}

// Resolve gathers bootstrap peers from every source. Seed servers are asked only when
// the other sources give fewer than want peers.
func Resolve(ctx context.Context, config []peer.AddrInfo, cachePath string, servers []string, want int) []peer.AddrInfo {
	out := Merge(append([]peer.AddrInfo(nil), config...), Shipped())
	if cachePath != "" {
		if cached, err := LoadCache(cachePath); err == nil {
			out = Merge(out, cached)
		}
	}
	for _, s := range servers {
		if len(out) >= want {
			break
		}
		if got, err := Fetch(ctx, s, want); err == nil {
			out = Merge(out, got)
		}
	}
	return out
}

type entry struct {
	PeerID string   `json:"peerId"`
	Addrs  []string `json:"addrs"`
}

// Fetch asks a seed server for up to n registered nodes.
func Fetch(ctx context.Context, server string, n int) ([]peer.AddrInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(server, "/")+"/peers?n="+fmt.Sprint(n), nil)
	if err != nil {
		return nil, err
	}
	var entries []entry
	if err := do(req, &entries); err != nil {
		return nil, err
	}
	var out []peer.AddrInfo
	for _, e := range entries {
		id, err := peer.Decode(e.PeerID)
		if err != nil {
			continue
		}
		ai := peer.AddrInfo{ID: id}
		for _, a := range e.Addrs {
			if m, err := ma.NewMultiaddr(a); err == nil {
				ai.Addrs = append(ai.Addrs, m)
			}
		}
		if len(ai.Addrs) > 0 {
			out = Merge(out, []peer.AddrInfo{ai})
		}
	}
	return out, nil
}

// RegisterMessage is what a node signs to register (must match librserver).
func RegisterMessage(nonce, peerID string, addrs []string) []byte {
	return []byte(strings.Join(append([]string{"libr-seed/v2", nonce, peerID}, addrs...), "\n"))
}

// Register publishes key's dialable addresses (loopback dropped, at most 8) on a seed
// server and returns how long the registration lives.
func Register(ctx context.Context, server string, key ed25519.PrivateKey, addrs []ma.Multiaddr) (time.Duration, error) {
	pub := key.Public().(ed25519.PublicKey)
	id, err := proto.PeerIDFor(pub)
	if err != nil {
		return 0, err
	}
	var list []string
	for _, a := range addrs {
		if !manet.IsIPLoopback(a) && len(list) < maxAddrs {
			list = append(list, a.String())
		}
	}
	if len(list) == 0 {
		return 0, errors.New("no dialable addresses to register")
	}
	base := strings.TrimRight(server, "/")
	pk := base64.StdEncoding.EncodeToString(pub)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/auth/challenge?publicKey="+url.QueryEscape(pk), nil)
	if err != nil {
		return 0, err
	}
	var ch struct{ Nonce string }
	if err := do(req, &ch); err != nil {
		return 0, fmt.Errorf("challenge: %w", err)
	}
	body, _ := json.Marshal(map[string]any{
		"publicKey": pk, "peerId": id, "addrs": list, "nonce": ch.Nonce,
		"signature": base64.StdEncoding.EncodeToString(ed25519.Sign(key, RegisterMessage(ch.Nonce, id, list))),
	})
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, base+"/nodes/register", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ok struct{ TTLSeconds int }
	if err := do(req, &ok); err != nil {
		return 0, fmt.Errorf("register: %w", err)
	}
	return time.Duration(ok.TTLSeconds) * time.Second, nil
}

func do(req *http.Request, v any) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxReply))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, bytes.TrimSpace(b))
	}
	return json.Unmarshal(b, v)
}

// LoadCache reads peers saved by SaveCache.
func LoadCache(path string) ([]peer.AddrInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return Parse(list), nil
}

// SaveCache stores up to 32 routing-table peers (DB nodes) with their known
// addresses, for bootstrapping the next run without a seed server.
func SaveCache(path string, h host.Host, d *dht.IpfsDHT) error {
	var list []string
	for _, p := range d.RoutingTable().ListPeers() {
		for _, a := range h.Peerstore().Addrs(p) {
			if manet.IsIPLoopback(a) {
				continue
			}
			list = append(list, a.Encapsulate(ma.StringCast("/p2p/"+p.String())).String())
		}
		if len(list) >= maxCached {
			break
		}
	}
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
