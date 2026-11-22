// Package p2p builds LIBR v2 libp2p hosts and the routing DHT. Every peer is Noise-
// secured end to end; NATed peers are reachable through circuit relay v2 and upgrade to
// direct connections with hole punching where the NAT allows it.
package p2p

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/x/rate"
)

// DHTPrefix keeps LIBR's DHT separate from the public IPFS DHT.
const DHTPrefix = protocol.ID("/libr")

// Config describes a host.
type Config struct {
	// Key is the app Ed25519 key. DB nodes and moderators set it so their peer ID is
	// derived from it; ordinary clients leave it nil (fresh peer ID per run).
	Key ed25519.PrivateKey
	// Listen multiaddrs, e.g. "/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/udp/4001/quic-v1".
	Listen []string
	// Relays this host reserves on when it is behind NAT.
	Relays []peer.AddrInfo
	// RelayService makes a publicly reachable node relay for others.
	RelayService bool
}

// BaseOptions are the options every LIBR host uses (production and tests).
func BaseOptions(cfg Config) ([]libp2p.Option, error) {
	rm, err := ResourceManager()
	if err != nil {
		return nil, err
	}
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Listen...),
		libp2p.Security(noise.ID, noise.New),
		libp2p.ResourceManager(rm),
		libp2p.EnableHolePunching(),
	}
	if cfg.Key != nil {
		k, err := crypto.UnmarshalEd25519PrivateKey(cfg.Key)
		if err != nil {
			return nil, err
		}
		opts = append(opts, libp2p.Identity(k))
	}
	if cfg.RelayService {
		opts = append(opts, libp2p.EnableRelayService(relayv2.WithResources(RelayResources())))
	}
	return opts, nil
}

// NewHost starts a host that keeps reservations on cfg.Relays while it is not
// publicly reachable.
func NewHost(cfg Config) (host.Host, error) {
	opts, err := BaseOptions(cfg)
	if err != nil {
		return nil, err
	}
	if len(cfg.Relays) > 0 {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(cfg.Relays))
	}
	return libp2p.New(append(opts, libp2p.NATPortMap())...)
}

// NewDHT starts the routing DHT. DB nodes are servers (they appear in routing tables,
// so the closest peers to any key are always DB nodes); clients and mods are clients.
func NewDHT(ctx context.Context, h host.Host, server bool, bootstrap []peer.AddrInfo) (*dht.IpfsDHT, error) {
	mode := dht.ModeClient
	if server {
		mode = dht.ModeServer
	}
	return dht.New(ctx, LimitedOK{h}, dht.Mode(mode), dht.ProtocolPrefix(DHTPrefix), dht.BootstrapPeers(bootstrap...))
}

// RelayResources are the limits a relay (stock relay or relaying DB node) enforces.
// Circuits are bounded so a relay cannot be used as a free bulk pipe; DHT and LIBR
// requests are small, and peers move to direct connections when hole punching works.
func RelayResources() relayv2.Resources {
	rc := relayv2.DefaultResources()
	rc.Limit = &relayv2.RelayLimit{Duration: 10 * time.Minute, Data: 8 << 20}
	rc.MaxCircuits = 64
	return rc
}

// ResourceManager is libp2p's default resource manager with a looser connection-rate
// limit. The default allows 0.2 new connections/s per IP after a burst of 16, and a
// relayed connection is charged to the relay's IP, so a peer that opens circuits to
// several mods and DB nodes through one relay gets "rate limit exceeded" (found by the
// Docker DHT spike; localhost is exempt, so in-process tests cannot show it).
func ResourceManager() (network.ResourceManager, error) {
	limiter := &rate.Limiter{SubnetRateLimiter: rate.SubnetLimiter{
		IPv4SubnetLimits: []rate.SubnetLimit{{PrefixLength: 32, Limit: rate.Limit{RPS: 20, Burst: 200}}},
		IPv6SubnetLimits: []rate.SubnetLimit{{PrefixLength: 56, Limit: rate.Limit{RPS: 20, Burst: 200}}},
		GracePeriod:      time.Minute,
	}}
	return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale()), rcmgr.WithConnRateLimiters(limiter))
}

// LimitedOK lets every outbound stream use a limited (relayed) connection. kad-dht does
// not opt in itself, and NATed peers are only reachable over circuits.
type LimitedOK struct{ host.Host }

func (h LimitedOK) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	return h.Host.NewStream(network.WithAllowLimitedConn(ctx, "libr"), p, pids...)
}
