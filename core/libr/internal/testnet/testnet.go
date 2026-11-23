// Package testnet builds in-process libp2p topologies for LIBR integration tests:
// public relays plus NATed peers that can only be reached through circuit relay v2.
package testnet

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/p2p"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
)

// Config describes a topology. All DB nodes and mods are NATed.
type Config struct {
	Relays int // public circuit relays
	DBs    int // DHT server nodes
	Mods   int // DHT client nodes that must stay findable by peer ID
}

// Node is a NATed peer with a DHT.
type Node struct {
	Host host.Host
	DHT  *dht.IpfsDHT
}

// Net is a running topology. Everything is closed by t.Cleanup.
type Net struct {
	Relays []host.Host
	DBs    []*Node
	Mods   []*Node
}

// New starts the topology and waits until every NATed peer holds a reservation on
// every relay and every DB node has joined the DHT.
func New(t testing.TB, cfg Config) *Net {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	n := &Net{}
	var relayInfos []peer.AddrInfo
	for i := 0; i < cfg.Relays; i++ {
		r, err := NewRelay("/ip4/127.0.0.1/tcp/0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { r.Close() })
		n.Relays = append(n.Relays, r)
		relayInfos = append(relayInfos, peer.AddrInfo{ID: r.ID(), Addrs: r.Addrs()})
	}

	natted := func(mode dht.ModeOpt) *Node {
		nd, err := NewNATed(ctx, "/ip4/127.0.0.1/tcp/0", relayInfos, mode)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { nd.Close() })
		return nd
	}
	for i := 0; i < cfg.DBs; i++ {
		n.DBs = append(n.DBs, natted(dht.ModeServer))
	}
	for i := 0; i < cfg.Mods; i++ {
		n.Mods = append(n.Mods, natted(dht.ModeClient))
	}

	// Bootstrap: everyone dials the first DB node through its circuit address.
	seed := n.DBs[0]
	for _, nd := range append(n.DBs[1:], n.Mods...) {
		if err := nd.Host.Connect(ctx, AddrInfo(seed.Host)); err != nil {
			t.Fatalf("bootstrap connect: %v", err)
		}
	}
	for _, nd := range append(n.DBs, n.Mods...) {
		if err := nd.DHT.Bootstrap(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, nd := range append(n.DBs, n.Mods...) {
		<-nd.DHT.RefreshRoutingTable()
	}
	for _, m := range n.Mods {
		if err := Announce(ctx, m); err != nil {
			t.Fatalf("announce: %v", err)
		}
	}
	return n
}

// AddrInfo returns h's advertised (circuit) addresses.
func AddrInfo(h host.Host) peer.AddrInfo {
	return peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
}

// NewRelay starts a public circuit relay v2 listening on listen.
func NewRelay(listen string, opts ...libp2p.Option) (host.Host, error) {
	rc := relayv2.DefaultResources()
	// DHT traffic rides relayed connections, so the default 2 min / 128 KiB cap is too small.
	rc.Limit = &relayv2.RelayLimit{Duration: 30 * time.Minute, Data: 1 << 24}
	// Every test peer shares 127.0.0.1, so the per-IP caps must not apply.
	rc.MaxReservationsPerIP, rc.MaxReservationsPerASN = 1<<10, 1<<10
	rc.MaxCircuits = 64
	return libp2p.New(append([]libp2p.Option{
		libp2p.ListenAddrStrings(listen),
		libp2p.ForceReachabilityPublic(),
		libp2p.EnableRelayService(relayv2.WithResources(rc)),
	}, opts...)...)
}

// NewNATed starts a peer that is reachable only through circuits on relays, holds a
// reservation on each, and runs the LIBR DHT in mode. opts are extra host options.
func NewNATed(ctx context.Context, listen string, relays []peer.AddrInfo, mode dht.ModeOpt, opts ...libp2p.Option) (*Node, error) {
	return NewNATedHost(ctx, p2p.Config{Listen: []string{listen}}, relays, mode == dht.ModeServer, opts...)
}

// NewNATedHost is NewNATed with the production host config (e.g. a stable key).
func NewNATedHost(ctx context.Context, cfg p2p.Config, relays []peer.AddrInfo, server bool, opts ...libp2p.Option) (*Node, error) {
	relaySet := map[peer.ID]bool{}
	for _, r := range relays {
		relaySet[r.ID] = true
	}
	cfg.Private, cfg.Relays = true, relays
	h, err := p2p.NewHost(cfg, append([]libp2p.Option{libp2p.ConnectionGater(natGater{relays: relaySet})}, opts...)...)
	if err != nil {
		return nil, err
	}
	if circuitAddrs(h.Addrs()) < len(relays) {
		h.Close()
		return nil, fmt.Errorf("%s advertises %v, want a circuit addr per relay", h.ID(), h.Addrs())
	}
	d, err := p2p.NewDHT(ctx, h, server, nil)
	if err != nil {
		h.Close()
		return nil, fmt.Errorf("dht: %w", err)
	}
	return &Node{Host: h, DHT: d}, nil
}

// Close stops the DHT and the host.
func (n *Node) Close() error {
	n.DHT.Close()
	return n.Host.Close()
}

// natGater models a symmetric NAT: outbound is free, inbound is accepted only from
// relays or over a relayed circuit. Hole punching therefore always fails, which is
// the worst case for routing.
type natGater struct{ relays map[peer.ID]bool }

func (natGater) InterceptPeerDial(peer.ID) bool               { return true }
func (natGater) InterceptAddrDial(peer.ID, ma.Multiaddr) bool { return true }
func (natGater) InterceptAccept(network.ConnMultiaddrs) bool  { return true }
func (g natGater) InterceptSecured(dir network.Direction, p peer.ID, cm network.ConnMultiaddrs) bool {
	return dir == network.DirOutbound || g.relays[p] || isCircuit(cm.RemoteMultiaddr())
}
func (natGater) InterceptUpgraded(network.Conn) (bool, control.DisconnectReason) { return true, 0 }

func isCircuit(a ma.Multiaddr) bool {
	_, err := a.ValueForProtocol(ma.P_CIRCUIT)
	return err == nil
}

func circuitAddrs(addrs []ma.Multiaddr) int {
	n := 0
	for _, a := range addrs {
		if isCircuit(a) {
			n++
		}
	}
	return n
}

// String is for test logs.
func (n *Node) String() string {
	return fmt.Sprintf("%s(rt=%d)", n.Host.ID().ShortString(), n.DHT.RoutingTable().Size())
}
