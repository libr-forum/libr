package dbnode

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libr-forum/Libr/core/libr/internal/testnet"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

const replicas = 4

// TestStoreOnClosestPeers: a NATed client locates the K DB nodes closest to a cert's
// hash through the DHT and gets a verified receipt from each, entirely over circuits.
func TestStoreOnClosestPeers(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 8 hosts")
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	reg, mods := community(t)

	relay, err := testnet.NewRelay("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { relay.Close() })
	relays := []peer.AddrInfo{testnet.AddrInfo(relay)}
	cfg := func(k []byte) p2p.Config { return p2p.Config{Key: k, Listen: []string{"/ip4/127.0.0.1/tcp/0"}} }

	var dbs []*testnet.Node
	for i := 0; i < 6; i++ {
		k := key(fmt.Sprint("db", i))
		n, err := testnet.NewNATedHost(ctx, cfg(k), relays, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { n.Close() })
		(&Node{Key: k, Sets: reg, Store: NewMemStore()}).Register(n.Host)
		dbs = append(dbs, n)
	}
	cl, err := testnet.NewNATedHost(ctx, cfg(nil), relays, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cl.Close() })
	for _, n := range append(dbs[1:], cl) {
		if err := n.Host.Connect(ctx, testnet.AddrInfo(dbs[0].Host)); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range append(dbs, cl) {
		<-n.DHT.RefreshRoutingTable()
	}

	c := cert(key("author"), "replicated", now, mods[0], mods[1], mods[2])
	h := c.Hash()
	closest, err := cl.DHT.GetClosestPeers(ctx, string(h[:]))
	if err != nil {
		t.Fatal(err)
	}
	if len(closest) < replicas {
		t.Fatalf("found %d peers, want >= %d", len(closest), replicas)
	}
	got := 0
	for _, p := range closest[:replicas] {
		var rr wire.ReceiptResp
		if err := wire.Call(ctx, cl.Host, p, wire.ProtoStore, c, &rr); err != nil {
			t.Fatalf("store on %s: %v", p, err)
		}
		if rr.Receipt == nil || !rr.Receipt.Verify() {
			t.Fatalf("store on %s: %+v", p, rr)
		}
		if id, _ := proto.PeerIDFor(rr.Receipt.Node); id != p.String() {
			t.Fatalf("receipt from %s signed by %s", p, id)
		}
		got++
	}
	if got != replicas {
		t.Fatalf("receipts %d/%d", got, replicas)
	}
}

// TestRelayCannotImpersonate: a relay (or anyone) that routes a circuit to the wrong
// peer fails the Noise handshake, because the dialer pins the peer ID it expects.
func TestRelayCannotImpersonate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	reg, _ := community(t)
	r := newRig(t, reg)

	relay, err := testnet.NewRelay("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatal(err)
	}
	defer relay.Close()
	impostor, err := testnet.NewNATedHost(ctx, p2p.Config{Listen: []string{"/ip4/127.0.0.1/tcp/0"}}, []peer.AddrInfo{testnet.AddrInfo(relay)}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer impostor.Close()

	// Ask for the real DB node's ID at an address that actually reaches the impostor.
	circuit := ma.StringCast(fmt.Sprintf("%s/p2p/%s/p2p-circuit", relay.Addrs()[0], relay.ID()))
	_ = r.client.Host.Network().ClosePeer(r.db)
	r.client.Host.Peerstore().ClearAddrs(r.db)
	err = r.client.Host.Connect(ctx, peer.AddrInfo{ID: r.db, Addrs: []ma.Multiaddr{circuit}})
	if err == nil {
		t.Fatal("connected to an impostor under the DB node's peer ID")
	}
}
