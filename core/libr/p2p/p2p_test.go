package p2p

import (
	"context"
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

// TestManyPeersBehindOneIP: more than libp2p's default 8 peers sharing one
// non-loopback IP can all connect to the same host. Skipped without such an address.
func TestManyPeersBehindOneIP(t *testing.T) {
	ip := lanIP(t)
	h, err := NewHost(Config{Listen: []string{"/ip4/" + ip + "/tcp/0"}})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	target := peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
	for i := 0; i < 20; i++ {
		c, err := libp2p.New(libp2p.NoListenAddrs)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if err := c.Connect(context.Background(), target); err != nil {
			t.Fatalf("peer %d behind %s: %v", i, ip, err)
		}
	}
	if n := len(h.Network().Peers()); n != 20 {
		t.Fatalf("%d peers connected, want 20", n)
	}
}

func lanIP(t *testing.T) string {
	t.Helper()
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	for _, a := range h.Addrs() {
		if ip, err := a.ValueForProtocol(4); err == nil && ip != "127.0.0.1" {
			return ip
		}
	}
	t.Skip(fmt.Sprint("no non-loopback IPv4 address: ", h.Addrs()))
	return ""
}
