// Command dhtspike runs one role of the v2 step 2 DHT spike in Docker: separate
// processes on isolated networks, joined only through circuit relays. See run.sh.
package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libr-forum/Libr/core/libr/internal/testnet"
)

const port = 4001

func main() {
	role := flag.String("role", "", "relay | db | mod | probe")
	name := flag.String("name", "", "node name; also its DNS name and key seed")
	relays := flag.String("relays", "relay-a,relay-b", "relay names")
	seed := flag.String("seed", "db1", "DB node everyone bootstraps from")
	dbs := flag.String("dbs", "", "all DB node names (probe ground truth)")
	mods := flag.String("mods", "", "mod names (probe)")
	k := flag.Int("k", 4, "replicas")
	keys := flag.Int("keys", 200, "probe: random keys to test")
	flag.Parse()

	ctx := context.Background()
	identity := libp2p.Identity(keyFor(*name))
	listen := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)

	if *role == "relay" {
		h, err := testnet.NewRelay(listen, identity)
		must(err)
		log.Printf("relay %s up", h.ID())
		select {}
	}

	var relayInfos []peer.AddrInfo
	for _, r := range split(*relays) {
		relayInfos = append(relayInfos, peer.AddrInfo{ID: idFor(r), Addrs: []ma.Multiaddr{ma.StringCast(fmt.Sprintf("/dns4/%s/tcp/%d", r, port))}})
	}
	mode := dht.ModeClient
	if *role == "db" {
		mode = dht.ModeServer
	}

	var n *testnet.Node
	retry(func() (err error) { n, err = testnet.NewNATed(ctx, listen, relayInfos, mode, identity); return err })
	if *role == "mod" {
		testnet.ServeEcho(n.Host)
	}
	if *name != *seed {
		seedInfo := peer.AddrInfo{ID: idFor(*seed)}
		for _, r := range relayInfos {
			seedInfo.Addrs = append(seedInfo.Addrs, ma.StringCast(fmt.Sprintf("%s/p2p/%s/p2p-circuit", r.Addrs[0], r.ID)))
		}
		retry(func() error { return n.Host.Connect(ctx, seedInfo) })
	}
	must(n.DHT.Bootstrap(ctx))
	log.Printf("%s %s joined; rt=%d", *role, n.Host.ID(), n.DHT.RoutingTable().Size())

	if *role == "mod" {
		for {
			if err := testnet.Announce(ctx, n); err != nil {
				log.Printf("announce: %v", err)
			}
			time.Sleep(30 * time.Second)
		}
	}
	if *role != "probe" {
		select {}
	}

	dbIDs := []peer.ID{}
	for _, d := range split(*dbs) {
		dbIDs = append(dbIDs, idFor(d))
	}
	// Let the routing table fill before measuring.
	for i := 0; i < 60 && n.DHT.RoutingTable().Size() < len(dbIDs)/2; i++ {
		<-n.DHT.RefreshRoutingTable()
		time.Sleep(time.Second)
	}
	if os.Getenv("SPIKE_DEBUG") != "" {
		debugMods(n, split(*mods))
		for _, m := range split(*mods) {
			log.Printf("DEBUG ReachPeer %s: %v", m, testnet.ReachPeer(n, idFor(m)))
		}
		debugMods(n, split(*mods))
		os.Exit(0)
	}
	failed := false
	rate, p95 := testnet.ClosestAccuracy(n, dbIDs, *k, *keys)
	fmt.Printf("RESULT closest-peers: %.1f%% exact top-%d over %d keys; p95 %v\n", rate*100, *k, *keys, p95)
	failed = failed || rate < 0.95 || p95 > 2*time.Second

	ok, total := 0, 0
	for round := 0; round < 25; round++ {
		for _, m := range split(*mods) {
			total++
			if err := testnet.ReachPeer(n, idFor(m)); err != nil {
				if total-ok <= 3 {
					log.Printf("reach %s: %v", m, err)
				}
			} else {
				ok++
			}
		}
	}
	fmt.Printf("RESULT find-mod: %d/%d\n", ok, total)
	failed = failed || total == 0 || float64(ok)/float64(total) < 0.99
	if failed {
		fmt.Println("RESULT FAIL")
		os.Exit(1)
	}
	fmt.Println("RESULT PASS")
}

// keyFor derives a stable Ed25519 key from a node name so peers know each other's IDs.
func keyFor(name string) crypto.PrivKey {
	seed := sha256.Sum256([]byte("libr-dht-spike/" + name))
	priv, _, err := crypto.GenerateEd25519Key(strings.NewReader(string(seed[:]) + string(seed[:])))
	must(err)
	return priv
}

func idFor(name string) peer.ID {
	id, err := peer.IDFromPrivateKey(keyFor(name))
	must(err)
	return id
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func retry(f func() error) {
	var err error
	for i := 0; i < 60; i++ {
		if err = f(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	log.Fatal(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// debugMods runs FindPeer for each mod and logs every DHT response it receives.
func debugMods(n *testnet.Node, mods []string) {
	for _, m := range mods {
		id := idFor(m)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ctx, events := routing.RegisterForQueryEvents(ctx)
		done := make(chan struct{})
		go func() {
			for ev := range events {
				if ev.Type != routing.PeerResponse {
					if ev.Type == routing.QueryError || ev.Type == routing.DialingPeer {
						log.Printf("DEBUG %s: event %d peer=%s extra=%s", m, ev.Type, ev.ID.ShortString(), ev.Extra)
					}
					continue
				}
				found := "absent"
				for _, pi := range ev.Responses {
					if pi.ID == id {
						found = fmt.Sprint(pi.Addrs)
					}
				}
				log.Printf("DEBUG %s: response from %s: mod %s (%d peers)", m, ev.ID.ShortString(), found, len(ev.Responses))
			}
			close(done)
		}()
		_, err := n.DHT.FindPeer(ctx, id)
		cancel()
		<-done
		log.Printf("DEBUG %s: FindPeer err=%v connectedness=%v addrs=%v", m, err, n.Host.Network().Connectedness(id), n.Host.Peerstore().Addrs(id))
		if err != nil {
			if sw, ok := n.Host.Network().(*swarm.Swarm); ok {
				sw.Backoff().Clear(id)
			}
			cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
			log.Printf("DEBUG %s: direct connect err=%v", m, n.Host.Connect(cctx, peer.AddrInfo{ID: id}))
			ccancel()
		}
	}
}
