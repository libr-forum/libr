package testnet

import (
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/libr-forum/Libr/core/libr/p2p"

	kb "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
)

// ClosestAccuracy asks asker for the k closest DB nodes to random keys and returns the
// fraction that exactly match the true XOR-closest k among all, plus p95 query latency.
func ClosestAccuracy(asker *Node, all []peer.ID, k, keys int) (rate float64, p95 time.Duration) {
	var lat []time.Duration
	exact := 0
	for i := 0; i < keys; i++ {
		key := fmt.Sprintf("spike-key-%d-%d", time.Now().UnixNano(), i)
		want := kb.SortClosestPeers(all, kb.ConvertKey(key))[:k]

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		t0 := time.Now()
		got, err := asker.DHT.GetClosestPeers(ctx, key)
		lat = append(lat, time.Since(t0))
		cancel()
		if err != nil || len(got) < k {
			continue
		}
		got = kb.SortClosestPeers(got, kb.ConvertKey(key))[:k]
		if sameSet(got, want) {
			exact++
		}
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	return float64(exact) / float64(keys), lat[len(lat)*95/100]
}

func sameSet(a, b []peer.ID) bool {
	m := map[peer.ID]bool{}
	for _, p := range a {
		m[p] = true
	}
	for _, p := range b {
		if !m[p] {
			return false
		}
	}
	return len(a) == len(b)
}

// EchoProto is served by ServeEcho; ReachPeer uses it to prove a stream works end to end.
const EchoProto = "/libr/spike/echo"

// ServeEcho echoes every stream on EchoProto.
func ServeEcho(h host.Host) {
	h.SetStreamHandler(EchoProto, func(s network.Stream) { _, _ = io.Copy(s, s); _ = s.Close() })
}

// ReachPeer resolves target via the DHT (dropping any cached address or connection
// first), dials it through a relay and round-trips one byte.
func ReachPeer(asker *Node, target peer.ID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = asker.Host.Network().ClosePeer(target) // force a real lookup
	asker.Host.Peerstore().ClearAddrs(target)
	// Each round must measure a fresh lookup+dial, not the backoff left by closing above.
	if sw, ok := asker.Host.Network().(*swarm.Swarm); ok {
		sw.Backoff().Clear(target)
	}

	info, err := asker.DHT.FindPeer(ctx, target)
	if err != nil {
		return err
	}
	if err := asker.Host.Connect(ctx, info); err != nil {
		return err
	}
	s, err := p2p.LimitedOK{Host: asker.Host}.NewStream(ctx, target, EchoProto)
	if err != nil {
		return err
	}
	defer s.Close()
	if _, err := s.Write([]byte{42}); err != nil {
		return err
	}
	_ = s.CloseWrite()
	b, err := io.ReadAll(s)
	if err != nil || len(b) != 1 || b[0] != 42 {
		return fmt.Errorf("echo failed: %v %v", b, err)
	}
	return nil
}

// Announce connects n to the DB nodes closest to its own peer ID and keeps those
// connections. Client-mode peers (mods) are not in anyone's routing table, so this is
// what lets FindPeer resolve them: those DB nodes hold n's circuit addresses.
func Announce(ctx context.Context, n *Node) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	closest, err := n.DHT.GetClosestPeers(ctx, string(n.Host.ID()))
	if err != nil {
		return err
	}
	for _, p := range closest {
		if err := n.Host.Connect(ctx, peer.AddrInfo{ID: p}); err == nil {
			n.Host.ConnManager().Protect(p, "libr-announce")
		}
	}
	return nil
}
