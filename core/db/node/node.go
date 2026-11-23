// Package node runs a LIBR v2 DB node: SQLite store, storage protocols, DHT server,
// joining through configured, shipped, cached and seed-server peers, and registering
// with seed servers. Used by the DB node binary and by the desktop app's "contribute
// storage" option.
package node

import (
	"context"
	"crypto/ed25519"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/db/sqlstore"
	"github.com/libr-forum/Libr/core/libr/dbnode"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/seeds"
)

// Config describes a DB node.
type Config struct {
	DataDir   string // node.key, db.sqlite, peers.json
	Sets      *proto.Registry
	Listen    []string
	Relays    []peer.AddrInfo // reserved on when Private
	Bootstrap []peer.AddrInfo
	SeedURLs  []string
	Public    bool // reachable: also serve as a circuit relay
	Private   bool // behind NAT: reserve on Relays, advertise circuit addresses only
	OptOut    []string
}

// Node is a running DB node.
type Node struct {
	Host  host.Host
	DHT   *dht.IpfsDHT
	store *sqlstore.Store

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Start brings the node up and returns; it joins the DHT and registers with seed
// servers in the background, retrying until Close.
func Start(cfg Config) (*Node, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, err
	}
	key, err := p2p.LoadOrCreateKey(filepath.Join(cfg.DataDir, "node.key"))
	if err != nil {
		return nil, err
	}
	h, err := p2p.NewHost(p2p.Config{Key: key, Listen: cfg.Listen, Relays: cfg.Relays, RelayService: cfg.Public, Private: cfg.Private})
	if err != nil {
		return nil, err
	}
	n := &Node{Host: h}
	fail := func(err error) (*Node, error) { n.Close(); return nil, err }
	if n.store, err = sqlstore.Open(filepath.Join(cfg.DataDir, "db.sqlite")); err != nil {
		return fail(err)
	}
	optOut := map[string]bool{}
	for _, c := range cfg.OptOut {
		optOut[c] = true
	}
	(&dbnode.Node{Key: key, Sets: cfg.Sets, Store: n.store, OptOut: optOut}).Register(h)

	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel
	if !cfg.Private {
		for _, r := range cfg.Relays {
			if err := h.Connect(ctx, r); err != nil {
				log.Printf("relay %s unreachable: %v", r.ID, err)
			}
		}
	}
	if n.DHT, err = p2p.NewDHT(ctx, h, true, cfg.Bootstrap); err != nil {
		return fail(err)
	}
	if err := n.DHT.Bootstrap(ctx); err != nil {
		return fail(err)
	}
	n.goRun(func() { join(ctx, h, n.DHT, cfg.Bootstrap, cfg.SeedURLs, filepath.Join(cfg.DataDir, "peers.json")) })
	for _, s := range cfg.SeedURLs {
		n.goRun(func() { register(ctx, h, key, s) })
	}
	log.Printf("dbnode %s serving %v; communities %v; opted out of %v", h.ID(), h.Addrs(), cfg.Sets.Communities(), cfg.OptOut)
	return n, nil
}

func (n *Node) goRun(f func()) {
	n.wg.Add(1)
	go func() { defer n.wg.Done(); f() }()
}

// Close stops the node and waits for its background work.
func (n *Node) Close() {
	if n.cancel != nil {
		n.cancel()
	}
	n.wg.Wait()
	if n.DHT != nil {
		n.DHT.Close()
	}
	if n.store != nil {
		n.store.Close()
	}
	n.Host.Close()
}

// join joins the DHT through the configured, shipped, cached and seed-server peers,
// retrying while none is reachable (peers start in any order), then caches the
// routing table for the next start.
func join(ctx context.Context, h host.Host, d *dht.IpfsDHT, boot []peer.AddrInfo, servers []string, cache string) {
	for {
		var peers []peer.AddrInfo
		for _, p := range seeds.Resolve(ctx, boot, cache, servers, 8) {
			if p.ID != h.ID() {
				peers = append(peers, p)
			}
		}
		if len(peers) > 0 {
			jctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			err := p2p.Join(jctx, h, d, peers)
			cancel()
			if err == nil {
				log.Printf("joined the DHT; routing table %d", d.RoutingTable().Size())
				if err := seeds.SaveCache(cache, h, d); err != nil {
					log.Printf("peer cache: %v", err)
				}
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

// register keeps this node listed on a seed server, renewing at half the TTL.
func register(ctx context.Context, h host.Host, key ed25519.PrivateKey, server string) {
	for {
		wait := 30 * time.Second
		ttl, err := seeds.Register(ctx, server, key, h.Addrs())
		if err != nil {
			log.Printf("seed %s: %v", server, err)
		} else if ttl/2 > wait {
			wait = ttl / 2
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}
