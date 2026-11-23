// Command relay is a LIBR v2 circuit relay: a stock libp2p circuit relay v2 with
// bounded circuits. It forwards encrypted bytes between NATed peers and never sees or
// routes application messages (v1's JSON proxy, peer map and server registration are
// gone). Any publicly reachable DB node can run the same service; clients keep
// reservations on several relays.
//
// Environment:
//
//	LIBP2P_PRIV_KEY  base64 protobuf libp2p key; stable peer ID (generated and printed if unset)
//	LIBR_LISTEN      comma-separated listen multiaddrs; default tcp+quic on 4001
//	RELAY_ANNOUNCE   comma-separated public multiaddrs to advertise instead of the
//	                 listen addresses (e.g. /dns4/example.org/tcp/443/wss behind a TLS proxy)
//	HEALTH_PORT      plain HTTP /health port; default 8080
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libr-forum/Libr/core/libr/p2p"
)

func main() {
	key, err := loadOrGenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	announce, err := multiaddrs(os.Getenv("RELAY_ANNOUNCE"))
	if err != nil {
		log.Fatalf("RELAY_ANNOUNCE: %v", err)
	}
	h, err := newRelay(key, split(env("LIBR_LISTEN", "/ip4/0.0.0.0/tcp/4001,/ip4/0.0.0.0/udp/4001/quic-v1")), announce)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()
	for _, a := range h.Addrs() {
		log.Printf("relay %s/p2p/%s", a, h.ID())
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })
		log.Print(http.ListenAndServe(":"+env("HEALTH_PORT", "8080"), mux))
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
}

// newRelay starts a relay host. A non-empty announce replaces the advertised addresses.
func newRelay(key libp2pcrypto.PrivKey, listen []string, announce []ma.Multiaddr) (host.Host, error) {
	opts, err := p2p.BaseOptions(p2p.Config{Listen: listen, RelayService: true})
	if err != nil {
		return nil, err
	}
	opts = append(opts, libp2p.Identity(key), libp2p.ForceReachabilityPublic())
	if len(announce) > 0 {
		opts = append(opts, libp2p.AddrsFactory(func([]ma.Multiaddr) []ma.Multiaddr { return announce }))
	}
	return libp2p.New(opts...)
}

// loadOrGenerateKey reads LIBP2P_PRIV_KEY or generates a key and prints it, so the
// operator can pin the relay's peer ID.
func loadOrGenerateKey() (libp2pcrypto.PrivKey, error) {
	if b64 := os.Getenv("LIBP2P_PRIV_KEY"); b64 != "" {
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("LIBP2P_PRIV_KEY: %w", err)
		}
		return libp2pcrypto.UnmarshalPrivateKey(data)
	}
	priv, _, err := libp2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	if b, err := libp2pcrypto.MarshalPrivateKey(priv); err == nil {
		log.Printf("LIBP2P_PRIV_KEY not set; this run's key (set it to keep the peer ID):\nLIBP2P_PRIV_KEY=%s", base64.StdEncoding.EncodeToString(b))
	}
	return priv, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
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

func multiaddrs(s string) ([]ma.Multiaddr, error) {
	var out []ma.Multiaddr
	for _, a := range split(s) {
		m, err := ma.NewMultiaddr(a)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
