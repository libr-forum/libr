// Package app is the LIBR v2 client node shared by the desktop (Wails) and mobile
// (gomobile) apps: it joins the network, posts, reads and removes per community, keeps
// posts and reports that wait for moderators and resumes them on a timer, and, when
// the user's key is in a community's moderator set, serves votes and a human review
// queue. Adapters only translate types.
package app

import (
	"context"
	"crypto/ed25519"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libr-forum/Libr/core/libr/forum"
	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/seeds"
)

// Shipped communities (verified genesis files). Production ones are added at release.
//
//go:embed communities
var shipped embed.FS

// Config describes a node. Empty lists fall back to LIBR_* environment variables.
type Config struct {
	DataDir   string             // pending items, peer cache, mod queue, extra communities
	Key       ed25519.PrivateKey // the user's identity (author, reporter, moderator)
	Relays    []string           // multiaddrs with /p2p/ID; LIBR_RELAYS
	Bootstrap []string           // LIBR_BOOTSTRAP
	SeedURLs  []string           // LIBR_SEED_URLS
	// Policy and Report decide for a moderator (return moderator.ErrDefer to queue an
	// item for human review). Nil: approve posts, queue every report.
	Policy moderator.Policy
	Report moderator.ReportPolicy
	// OnFinalized is called when a pending post or report completes, is rejected or
	// expires.
	OnFinalized func(PendingInfo)
	// Communities overrides the shipped set (tests); nil uses shipped + DataDir +
	// LIBR_COMMUNITIES.
	Communities []proto.CommunityFile
}

// Community is what the UI lists.
type Community struct {
	Name      string   `json:"name"`
	IsMod     bool     `json:"is_mod"`
	Mods      int      `json:"mods"`
	Threshold int      `json:"threshold"`
	ModKeys   []string `json:"mod_keys"` // base64 Ed25519 keys of the moderator set
}

// Post is one verified post.
type Post struct {
	Hash      string `json:"hash"`
	MsgID     string `json:"msg_id"`
	Community string `json:"community"`
	Author    string `json:"author"` // base64 Ed25519 key
	Ts        int64  `json:"ts"`
	Content   string `json:"content"`
	Votes     int    `json:"votes"`
	Mine      bool   `json:"mine"`
}

// Outcome of Post and Report.
type Outcome struct {
	Status string `json:"status"` // posted | removed | pending | rejected
	Hash   string `json:"hash,omitempty"`
	Ts     int64  `json:"ts,omitempty"`
	ID     string `json:"id,omitempty"` // pending item, while status is pending
	Error  string `json:"error,omitempty"`
}

// App is a running node.
type App struct {
	cfg    Config
	host   host.Host
	dht    *dht.IpfsDHT
	reg    *proto.Registry
	client *forum.Client
	mod    *moderator.Mod
	store  *pendingStore
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.Mutex
	certs  map[proto.Hash]proto.MsgCert // seen in feeds and own posts, for report/delete
	tickMu sync.Mutex
}

// Tick is how often pending items are resumed.
const Tick = 30 * time.Second

// Start joins the network and starts the pending-item timer (and moderator duties
// if the key is in a moderator set).
func Start(ctx context.Context, cfg Config) (*App, error) {
	if cfg.Key == nil {
		return nil, errors.New("app: no key")
	}
	if cfg.DataDir == "" {
		return nil, errors.New("app: no data dir")
	}
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, err
	}
	reg, err := loadCommunities(cfg)
	if err != nil {
		return nil, err
	}
	envList(&cfg.Relays, "LIBR_RELAYS")
	envList(&cfg.Bootstrap, "LIBR_BOOTSTRAP")
	envList(&cfg.SeedURLs, "LIBR_SEED_URLS")
	relays := seeds.Parse(cfg.Relays)
	pub := cfg.Key.Public().(ed25519.PublicKey)
	isMod := false
	for _, c := range reg.Communities() {
		if s, _ := reg.Head(c); s.IsMember(pub) {
			isMod = true
		}
	}

	hcfg := p2p.Config{Listen: []string{"/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/udp/0/quic-v1"}, Relays: relays, Private: len(relays) > 0}
	if isMod {
		hcfg.Key = cfg.Key // moderators must be findable under the peer ID in the ModSet
	}
	h, err := p2p.NewHost(hcfg)
	if err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a := &App{cfg: cfg, host: h, reg: reg, cancel: cancel, certs: map[proto.Hash]proto.MsgCert{}}
	fail := func(err error) (*App, error) { a.Close(); return nil, err }
	if a.dht, err = p2p.NewDHT(runCtx, h, false, nil); err != nil {
		return fail(err)
	}
	a.client = &forum.Client{Host: h, DHT: a.dht, Sets: reg, Key: cfg.Key}
	if a.store, err = openPending(filepath.Join(cfg.DataDir, "pending")); err != nil {
		return fail(err)
	}
	if isMod {
		q, err := moderator.NewQueue(filepath.Join(cfg.DataDir, "modqueue.json"))
		if err != nil {
			return fail(err)
		}
		a.mod = &moderator.Mod{Key: cfg.Key, Sets: reg, Policy: cfg.Policy, Report: cfg.Report, Queue: q}
		if a.mod.Policy == nil {
			a.mod.Policy = moderator.ApproveAll
		}
		if a.mod.Report == nil {
			a.mod.Report = func(context.Context, proto.Msg, string) (bool, error) { return false, moderator.ErrDefer }
		}
		a.mod.Register(h)
	}

	boot, err := infos(cfg.Bootstrap)
	if err != nil {
		return fail(err)
	}
	cache := filepath.Join(cfg.DataDir, "peers.json")
	jctx, jcancel := context.WithTimeout(ctx, 30*time.Second)
	defer jcancel()
	if err := p2p.Join(jctx, h, a.dht, seeds.Resolve(jctx, boot, cache, cfg.SeedURLs, 8)); err != nil {
		return fail(fmt.Errorf("join: %w", err))
	}
	_ = seeds.SaveCache(cache, h, a.dht)
	if isMod {
		a.go_(func() { p2p.KeepFindable(runCtx, h, a.dht, 30*time.Second) })
	}
	a.go_(func() {
		t := time.NewTicker(Tick)
		defer t.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-t.C:
				a.Tick(runCtx)
			}
		}
	})
	return a, nil
}

func (a *App) go_(f func()) {
	a.wg.Add(1)
	go func() { defer a.wg.Done(); f() }()
}

// Close stops the node.
func (a *App) Close() {
	a.cancel()
	a.wg.Wait()
	if a.dht != nil {
		a.dht.Close()
	}
	a.host.Close()
}

// Communities lists known communities and whether the user moderates each.
func (a *App) Communities() []Community {
	pub := a.cfg.Key.Public().(ed25519.PublicKey)
	var out []Community
	for _, name := range a.reg.Communities() {
		s, _ := a.reg.Head(name)
		c := Community{Name: name, IsMod: s.IsMember(pub), Mods: len(s.Mods), Threshold: int(s.Threshold)}
		for _, m := range s.Mods {
			c.ModKeys = append(c.ModKeys, b64(m.PubKey))
		}
		out = append(out, c)
	}
	return out
}

// IsMod reports whether the user is in community's current moderator set.
func (a *App) IsMod(community string) bool {
	s, ok := a.reg.Head(community)
	return ok && s.IsMember(a.cfg.Key.Public().(ed25519.PublicKey))
}

// Post publishes content. A post that moderators are reviewing, or that could not
// reach enough of them, is kept and retried every Tick.
func (a *App) Post(ctx context.Context, community, content string) (Outcome, error) {
	posted, err := a.client.Submit(ctx, community, []byte(content))
	return a.outcome(posted, nil, err, func(p *forum.Pending) (string, error) { return a.store.add(p, content) })
}

// Report asks moderators to remove the post with hash (hex).
func (a *App) Report(ctx context.Context, hash, reason string) (Outcome, error) {
	cert, err := a.cert(ctx, hash)
	if err != nil {
		return Outcome{}, err
	}
	removed, err := a.client.Report(ctx, cert, reason)
	return a.outcome(nil, removed, err, func(p *forum.Pending) (string, error) { return a.store.add(p, "") })
}

// Delete withdraws the user's own post.
func (a *App) Delete(ctx context.Context, hash string) error {
	cert, err := a.cert(ctx, hash)
	if err != nil {
		return err
	}
	r, err := a.client.Delete(ctx, cert)
	if err == nil {
		go func() { <-r.Done }() // background acks finish on their own
	}
	return err
}

func (a *App) outcome(posted *forum.Posted, removed *forum.Removed, err error, keep func(*forum.Pending) (string, error)) (Outcome, error) {
	var qe *forum.QuorumError
	switch {
	case errors.As(err, &qe):
		id, kerr := keep(qe.Pending)
		if kerr != nil {
			return Outcome{}, kerr
		}
		return Outcome{Status: "pending", ID: id, Ts: qe.Pending.Cert.Msg.Ts}, nil
	case errors.Is(err, forum.ErrRejected):
		return Outcome{Status: "rejected"}, nil
	case err != nil:
		return Outcome{}, err
	case posted != nil:
		a.remember(posted.Cert)
		return Outcome{Status: "posted", Hash: hex.EncodeToString(posted.Hash[:]), Ts: posted.Cert.Msg.Ts}, nil
	}
	return Outcome{Status: "removed"}, nil
}

// Feed returns community's verified posts from the last `since`, newest first.
func (a *App) Feed(ctx context.Context, community string, since time.Duration) ([]Post, error) {
	now := time.Now()
	certs, err := a.client.Feed(ctx, community, now.Add(-since), now)
	if err != nil {
		return nil, err
	}
	pub := a.cfg.Key.Public().(ed25519.PublicKey)
	out := make([]Post, 0, len(certs))
	for _, c := range certs {
		a.remember(c)
		out = append(out, toPost(c, pub))
	}
	return out, nil
}

func toPost(c proto.MsgCert, me ed25519.PublicKey) Post {
	h, id := c.Hash(), c.Msg.ID()
	return Post{
		Hash: hex.EncodeToString(h[:]), MsgID: hex.EncodeToString(id[:]), Community: c.Msg.Community,
		Author: b64(c.Msg.Author), Ts: c.Msg.Ts, Content: string(c.Msg.Content), Votes: len(c.Votes),
		Mine: string(c.Msg.Author) == string(me),
	}
}

func (a *App) remember(c proto.MsgCert) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.certs) > 10_000 {
		a.certs = map[proto.Hash]proto.MsgCert{} // shortcut: a cache, refetched on demand
	}
	a.certs[c.Hash()] = c
}

func (a *App) cert(ctx context.Context, hash string) (proto.MsgCert, error) {
	var h proto.Hash
	if b, err := hex.DecodeString(hash); err != nil || len(b) != len(h) {
		return proto.MsgCert{}, fmt.Errorf("bad hash %q", hash)
	} else {
		copy(h[:], b)
	}
	a.mu.Lock()
	c, ok := a.certs[h]
	a.mu.Unlock()
	if ok {
		return c, nil
	}
	return a.client.Get(ctx, h)
}

// PeerID is this node's libp2p peer ID.
func (a *App) PeerID() string { return a.host.ID().String() }

// Status is the network state for the UI.
type Status struct {
	Online       bool `json:"online"`
	Peers        int  `json:"peers"`
	RoutingTable int  `json:"routing_table"`
	Reservations int  `json:"reservations"` // relays currently holding a reservation
}

func (a *App) Status() Status {
	s := Status{Peers: len(a.host.Network().Peers()), RoutingTable: a.dht.RoutingTable().Size()}
	for _, addr := range a.host.Addrs() {
		if _, err := addr.ValueForProtocol(ma.P_CIRCUIT); err == nil {
			s.Reservations++
		}
	}
	s.Online = s.RoutingTable > 0
	return s
}

// ModQueue lists the human review queue (nil if the user moderates nothing).
func (a *App) ModQueue() []moderator.Item {
	if a.mod == nil {
		return nil
	}
	return a.mod.Queue.Items()
}

// ModDecide records the user's decision on a queued item.
func (a *App) ModDecide(key string, approve bool) error {
	if a.mod == nil {
		return errors.New("not a moderator")
	}
	return a.mod.Resolve(key, approve)
}

func loadCommunities(cfg Config) (*proto.Registry, error) {
	reg := &proto.Registry{}
	if cfg.Communities != nil {
		for _, cf := range cfg.Communities {
			if err := reg.Add(cf); err != nil {
				return nil, err
			}
		}
		return reg, nil
	}
	// First definition wins: LIBR_COMMUNITIES, then the data dir, then shipped.
	sub, _ := fs.Sub(shipped, "communities")
	sources := []fs.FS{}
	if d := os.Getenv("LIBR_COMMUNITIES"); d != "" {
		sources = append(sources, os.DirFS(d))
	}
	sources = append(sources, os.DirFS(filepath.Join(cfg.DataDir, "communities")), sub)
	for _, src := range sources {
		names, _ := fs.Glob(src, "*.json")
		for _, n := range names {
			b, err := fs.ReadFile(src, n)
			if err != nil {
				continue
			}
			var cf proto.CommunityFile
			if err := json.Unmarshal(b, &cf); err != nil {
				return nil, fmt.Errorf("community %s: %w", n, err)
			}
			if len(cf.Chain) > 0 {
				if _, dup := reg.Head(cf.Chain[0].Community); dup {
					continue
				}
			}
			if err := reg.Add(cf); err != nil {
				return nil, fmt.Errorf("community %s: %w", n, err)
			}
		}
	}
	return reg, nil
}

func envList(dst *[]string, key string) {
	if len(*dst) == 0 {
		for _, s := range strings.Split(os.Getenv(key), ",") {
			if s = strings.TrimSpace(s); s != "" {
				*dst = append(*dst, s)
			}
		}
	}
}

func infos(list []string) ([]peer.AddrInfo, error) {
	var out []peer.AddrInfo
	for _, s := range list {
		ai, err := peer.AddrInfoFromString(s)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", s, err)
		}
		out = append(out, *ai)
	}
	return out, nil
}

func sortedInfos(list []PendingInfo) []PendingInfo {
	sort.Slice(list, func(i, j int) bool { return list[i].Created > list[j].Created })
	return list
}
