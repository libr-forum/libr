package testnet

// Spike (v2 step 2): does go-libp2p-kad-dht route correctly when every DB node is
// NATed and reachable only through circuit relays? Results feed research/06-dht-spike.md.

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	nDBs      = 12
	nMods     = 4
	replicasK = 4 // LIBR stores each cert on the K closest DB nodes
)

func TestDHTOverCircuitRelays(t *testing.T) {
	if testing.Short() {
		t.Skip("spike: starts 18 hosts")
	}
	n := New(t, Config{Relays: 2, DBs: nDBs, Mods: nMods})
	for _, m := range n.Mods {
		ServeEcho(m.Host)
	}

	t.Run("routing tables hold DB peers over circuits only", func(t *testing.T) {
		for _, d := range n.DBs {
			if size := d.DHT.RoutingTable().Size(); size < nDBs/2 {
				t.Errorf("%s: routing table has %d peers, want >= %d", d, size, nDBs/2)
			}
			for _, c := range d.Host.Network().Conns() {
				if isDB(n, c.RemotePeer()) && !isCircuit(c.RemoteMultiaddr()) {
					t.Errorf("direct DB-DB connection %s -> %s; NAT model broken", d.Host.ID(), c.RemotePeer())
				}
			}
		}
	})

	t.Run("GetClosestPeers matches ground truth", func(t *testing.T) {
		rate, p95 := ClosestAccuracy(n.Mods[0], dbIDs(n), replicasK, 200)
		t.Logf("RESULT closest-peers: %.1f%% exact top-%d over 200 keys; p95 latency %v", rate*100, replicasK, p95)
		if rate < 0.95 {
			t.Errorf("accuracy %.1f%% < 95%%", rate*100)
		}
		if p95 > 2*time.Second {
			t.Errorf("p95 %v > 2s", p95)
		}
	})

	t.Run("FindPeer locates client-mode mods and streams reach them", func(t *testing.T) {
		ok, total := 0, 0
		for round := 0; round < 25; round++ {
			for i, target := range n.Mods {
				total++
				asker := n.Mods[(i+1)%len(n.Mods)]
				if ReachPeer(asker, target.Host.ID()) == nil {
					ok++
				}
			}
		}
		rate := float64(ok) / float64(total)
		t.Logf("RESULT find-mod: %d/%d (%.1f%%)", ok, total, rate*100)
		if rate < 0.99 {
			t.Errorf("FindPeer+stream success %.1f%% < 99%%", rate*100)
		}
	})

	t.Run("survives losing one of two relays", func(t *testing.T) {
		start := time.Now()
		_ = n.Relays[0].Close()
		var rate float64
		for time.Since(start) < 60*time.Second {
			if rate, _ = ClosestAccuracy(n.Mods[0], dbIDs(n), replicasK, 50); rate >= 0.95 {
				break
			}
			time.Sleep(2 * time.Second)
		}
		t.Logf("RESULT relay-failover: %.1f%% accuracy %v after killing relay", rate*100, time.Since(start).Round(time.Second))
		if rate < 0.95 {
			t.Errorf("accuracy after relay loss %.1f%% < 95%% within 60s", rate*100)
		}
	})
}

func dbIDs(n *Net) []peer.ID {
	ids := make([]peer.ID, len(n.DBs))
	for i, d := range n.DBs {
		ids[i] = d.Host.ID()
	}
	return ids
}

func isDB(n *Net, p peer.ID) bool {
	for _, d := range n.DBs {
		if d.Host.ID() == p {
			return true
		}
	}
	return false
}
