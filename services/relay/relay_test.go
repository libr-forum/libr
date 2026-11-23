package main

import (
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
)

// TestRelaysOpaqueStream: two peers that cannot dial each other exchange bytes through
// the relay over a circuit. The relay only forwards; it has no application handler.
func TestRelaysOpaqueStream(t *testing.T) {
	ctx := context.Background()
	key, _, _ := libp2pcrypto.GenerateEd25519Key(rand.Reader)
	relay, err := newRelay(key, []string{"/ip4/127.0.0.1/tcp/0"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer relay.Close()
	if n := len(relay.Mux().Protocols()); n == 0 {
		t.Fatal("no protocols")
	}
	for _, p := range relay.Mux().Protocols() {
		if p == "/chat/1.0.0" {
			t.Fatal("v1 application proxy still registered")
		}
	}
	ri := peer.AddrInfo{ID: relay.ID(), Addrs: relay.Addrs()}

	// Listener: no listen addresses, so the only way in is the relay.
	b, err := libp2p.New(libp2p.NoListenAddrs, libp2p.EnableRelay())
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := b.Connect(ctx, ri); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Reserve(ctx, b, ri); err != nil {
		t.Fatal(err)
	}
	b.SetStreamHandler("/test/echo", func(s network.Stream) { _, _ = io.Copy(s, s); _ = s.Close() })

	a, err := libp2p.New(libp2p.NoListenAddrs, libp2p.EnableRelay())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	circuit := ma.StringCast(relay.Addrs()[0].String() + "/p2p/" + relay.ID().String() + "/p2p-circuit")
	if err := a.Connect(ctx, peer.AddrInfo{ID: b.ID(), Addrs: []ma.Multiaddr{circuit}}); err != nil {
		t.Fatal(err)
	}
	s, err := a.NewStream(network.WithAllowLimitedConn(ctx, "test"), b.ID(), "/test/echo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	_ = s.CloseWrite()
	got, err := io.ReadAll(s)
	if err != nil || string(got) != "ping" {
		t.Fatalf("echo through relay: %q %v", got, err)
	}
}
