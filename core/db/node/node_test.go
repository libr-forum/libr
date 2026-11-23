package node

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// TestTwoNodesJoinAndServe: a second node joins the DHT through the first, both serve
// the storage protocols, the key persists across restarts, and Close stops cleanly.
func TestTwoNodesJoinAndServe(t *testing.T) {
	reg := &proto.Registry{}
	listen := []string{"/ip4/127.0.0.1/tcp/0"}
	a, err := Start(Config{DataDir: t.TempDir(), Sets: reg, Listen: listen})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	dirB := t.TempDir()
	boot := []peer.AddrInfo{{ID: a.Host.ID(), Addrs: a.Host.Addrs()}}
	b, err := Start(Config{DataDir: dirB, Sets: reg, Listen: listen, Bootstrap: boot})
	if err != nil {
		t.Fatal(err)
	}
	idB := b.Host.ID()
	for start := time.Now(); b.DHT.RoutingTable().Size() == 0; time.Sleep(50 * time.Millisecond) {
		if time.Since(start) > 10*time.Second {
			t.Fatal("second node never joined the DHT")
		}
	}

	// Serves the protocols: a cert for an unknown community is refused, not ignored.
	_, priv, _ := ed25519.GenerateKey(nil)
	c := proto.SignMsg(priv, proto.Msg{Community: "nope", Ts: time.Now().Unix(), Content: []byte("x")})
	var rr wire.ReceiptResp
	if err := wire.Call(context.Background(), a.Host, idB, wire.ProtoStore, c, &rr); err != nil || rr.Receipt != nil || rr.Err == "" {
		t.Fatalf("store on second node: %+v %v", rr, err)
	}

	b.Close()
	b2, err := Start(Config{DataDir: dirB, Sets: reg, Listen: listen})
	if err != nil {
		t.Fatal(err)
	}
	defer b2.Close()
	if b2.Host.ID() != idB {
		t.Fatal("node key did not persist across restarts")
	}
}
