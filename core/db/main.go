// Command dbnode runs a LIBR v2 storage node. Configuration is by environment:
//
//	LIBR_DATA_DIR     data directory (node key, SQLite); default ~/.config/libr/v2/db
//	LIBR_COMMUNITIES  directory of community *.json files (verified moderator sets)
//	LIBR_LISTEN       comma-separated listen multiaddrs; default tcp+quic on 4001
//	LIBR_RELAYS       comma-separated relay multiaddrs (with /p2p/<id>) to reserve on
//	LIBR_BOOTSTRAP    comma-separated multiaddrs of DHT peers to join through
//	LIBR_PUBLIC       "1" if publicly reachable: also serve as a circuit relay
//	LIBR_PRIVATE      "1" if behind NAT: reserve on LIBR_RELAYS, advertise circuit addresses only
//	LIBR_SEED_URLS    comma-separated seed servers (librserver): register this node, find peers
//	LIBR_OPTOUT       comma-separated communities this node will not store
package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dir := env("LIBR_DATA_DIR", defaultDataDir())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	key, err := p2p.LoadOrCreateKey(filepath.Join(dir, "node.key"))
	if err != nil {
		return err
	}

	commDir := os.Getenv("LIBR_COMMUNITIES")
	if commDir == "" {
		return errors.New("LIBR_COMMUNITIES is required: without moderator sets every certificate is refused")
	}
	reg, err := proto.LoadRegistry(os.DirFS(commDir))
	if err != nil {
		return fmt.Errorf("communities: %w", err)
	}

	relays, err := addrInfos(os.Getenv("LIBR_RELAYS"))
	if err != nil {
		return fmt.Errorf("LIBR_RELAYS: %w", err)
	}
	boot, err := addrInfos(os.Getenv("LIBR_BOOTSTRAP"))
	if err != nil {
		return fmt.Errorf("LIBR_BOOTSTRAP: %w", err)
	}

	h, err := p2p.NewHost(p2p.Config{
		Key:          key,
		Listen:       split(env("LIBR_LISTEN", "/ip4/0.0.0.0/tcp/4001,/ip4/0.0.0.0/udp/4001/quic-v1")),
		Relays:       relays,
		RelayService: os.Getenv("LIBR_PUBLIC") == "1",
		Private:      os.Getenv("LIBR_PRIVATE") == "1",
	})
	if err != nil {
		return err
	}
	defer h.Close()

	store, err := sqlstore.Open(filepath.Join(dir, "db.sqlite"))
	if err != nil {
		return err
	}
	defer store.Close()
	optOut := map[string]bool{}
	for _, c := range split(os.Getenv("LIBR_OPTOUT")) {
		optOut[c] = true
	}
	(&dbnode.Node{Key: key, Sets: reg, Store: store, OptOut: optOut}).Register(h)
	servers := split(os.Getenv("LIBR_SEED_URLS"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	for _, r := range relays {
		if err := h.Connect(ctx, r); err != nil {
			log.Printf("relay %s unreachable: %v", r.ID, err)
		}
	}
	d, err := p2p.NewDHT(ctx, h, true, boot)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Bootstrap(ctx); err != nil {
		return err
	}
	go join(ctx, h, d, boot, servers, filepath.Join(dir, "peers.json"))
	for _, s := range servers {
		go register(ctx, h, key, s)
	}

	log.Printf("dbnode %s serving %v; communities %v; opted out of %v", h.ID(), h.Addrs(), reg.Communities(), split(os.Getenv("LIBR_OPTOUT")))
	<-ctx.Done()
	return nil
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

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func defaultDataDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "libr", "v2", "db")
	}
	return "libr-db"
}

func split(s string) []string {
	var out []string
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func addrInfos(s string) ([]peer.AddrInfo, error) {
	var out []peer.AddrInfo
	for _, a := range split(s) {
		ai, err := peer.AddrInfoFromString(a)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", a, err)
		}
		out = append(out, *ai)
	}
	return out, nil
}
