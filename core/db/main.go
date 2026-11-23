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
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/db/node"
	"github.com/libr-forum/Libr/core/libr/proto"
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
	n, err := node.Start(node.Config{
		DataDir:   dir,
		Sets:      reg,
		Listen:    split(env("LIBR_LISTEN", "/ip4/0.0.0.0/tcp/4001,/ip4/0.0.0.0/udp/4001/quic-v1")),
		Relays:    relays,
		Bootstrap: boot,
		SeedURLs:  split(os.Getenv("LIBR_SEED_URLS")),
		Public:    os.Getenv("LIBR_PUBLIC") == "1",
		Private:   os.Getenv("LIBR_PRIVATE") == "1",
		OptOut:    split(os.Getenv("LIBR_OPTOUT")),
	})
	if err != nil {
		return err
	}
	defer n.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
	return nil
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
