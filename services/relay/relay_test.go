package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func newHost(t *testing.T) host.Host {
	t.Helper()
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.Close() })
	return h
}

// sendRegister opens a /chat/1.0.0 stream from c to relay and registers as claimed.
func sendRegister(t *testing.T, c, relay host.Host, claimed string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Connect(ctx, peer.AddrInfo{ID: relay.ID(), Addrs: relay.Addrs()}); err != nil {
		t.Fatal(err)
	}
	s, err := c.NewStream(ctx, relay.ID(), "/chat/1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(reqFormat{Type: "register", PeerID: claimed})
	if _, err := s.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := s.CloseWrite(); err != nil {
		t.Fatal(err)
	}
	// Wait for the relay to finish handling the request and close the stream (EOF).
	_, _ = s.Read(make([]byte, 1))
	_ = s.Close()
}

func lookup(id string) string {
	mu.RLock()
	defer mu.RUnlock()
	return IDmap[id]
}

func TestRegisterBindsToAuthenticatedPeer(t *testing.T) {
	relay := newHost(t)
	relay.SetStreamHandler("/chat/1.0.0", handleChatStream)
	victim, attacker := newHost(t), newHost(t)

	sendRegister(t, attacker, relay, victim.ID().String())
	if got := lookup(victim.ID().String()); got != "" {
		t.Fatalf("attacker hijacked victim's ID: IDmap[victim]=%s", got)
	}

	sendRegister(t, victim, relay, victim.ID().String())
	if got := lookup(victim.ID().String()); got != victim.ID().String() {
		t.Fatalf("honest register: IDmap[victim]=%q", got)
	}

	// A later spoof must not overwrite the honest entry.
	sendRegister(t, attacker, relay, victim.ID().String())
	if got := lookup(victim.ID().String()); got != victim.ID().String() {
		t.Fatalf("spoof overwrote entry: IDmap[victim]=%q", got)
	}
}
