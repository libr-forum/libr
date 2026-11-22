// Command librctl is the headless LIBR v2 tool: key and genesis generation, a
// moderator daemon, and a posting/reading client. Used by the testbed and for
// measurements (--trace writes one JSON line per network request).
//
//	librctl keygen  -out FILE [-libp2p]
//	librctl genesis -community NAME -f 1 -founders KEYFILE,... -mods PUBKEY,...
//	librctl mod serve -key FILE -communities DIR -policy approve-all|keyword:w1,w2 [net flags]
//	librctl client submit -communities DIR -community NAME -text TEXT [net flags]
//	librctl client feed   -communities DIR -community NAME [-since 1h] [-json] [net flags]
//
// Net flags: -relays, -bootstrap (comma-separated multiaddrs with /p2p/ID), -listen,
// -communities, -private, -trace FILE. The first four default to the DB node's
// environment variables LIBR_RELAYS, LIBR_BOOTSTRAP, LIBR_LISTEN, LIBR_COMMUNITIES.
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/forum"
	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
)

func main() {
	log.SetFlags(log.Ltime)
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "librctl:", err)
		os.Exit(1)
	}
}

const usage = "usage: librctl keygen | genesis | mod serve | client submit | client feed (see -h of each)"

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	cmd, args := args[0], args[1:]
	if (cmd == "mod" || cmd == "client") && len(args) > 0 {
		cmd, args = cmd+" "+args[0], args[1:]
	}
	switch cmd {
	case "keygen":
		return keygen(args, out)
	case "genesis":
		return genesis(args, out)
	case "mod serve":
		return modServe(args)
	case "client submit", "client feed":
		return clientCmd(cmd, args, out)
	}
	return errors.New(usage)
}

func keygen(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
	path := fs.String("out", "", "seed file to create (or read, if it exists)")
	libp2pKey := fs.Bool("libp2p", false, "also print the key as LIBP2P_PRIV_KEY (relay format)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("keygen: -out is required")
	}
	k, err := p2p.LoadOrCreateKey(*path)
	if err != nil {
		return err
	}
	pub := k.Public().(ed25519.PublicKey)
	id, err := proto.PeerIDFor(pub)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s %s", id, base64.StdEncoding.EncodeToString(pub))
	if *libp2pKey {
		lk, err := crypto.UnmarshalEd25519PrivateKey(k)
		if err != nil {
			return err
		}
		b, err := crypto.MarshalPrivateKey(lk)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, " %s", base64.StdEncoding.EncodeToString(b))
	}
	fmt.Fprintln(out)
	return nil
}

// genesis prints a CommunityFile: epoch 0 of community, signed by every founder.
func genesis(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("genesis", flag.ContinueOnError)
	name := fs.String("community", "", "community name")
	f := fs.Uint64("f", 1, "faulty moderators tolerated (needs >= 3f+1 mods)")
	founders := fs.String("founders", "", "comma-separated founder seed files")
	mods := fs.String("mods", "", "comma-separated moderator public keys (base64)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s := proto.ModSet{Community: *name, F: *f, Threshold: 2**f + 1}
	for _, b64 := range split(*mods) {
		pk, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(pk) != ed25519.PublicKeySize {
			return fmt.Errorf("mod key %q: not a base64 Ed25519 public key", b64)
		}
		id, err := proto.PeerIDFor(pk)
		if err != nil {
			return err
		}
		s.Mods = append(s.Mods, proto.ModEntry{PubKey: pk, PeerID: id})
	}
	var fpubs []ed25519.PublicKey
	for _, path := range split(*founders) {
		k, err := readKey(path)
		if err != nil {
			return err
		}
		proto.SignModSet(k, &s)
		fpubs = append(fpubs, k.Public().(ed25519.PublicKey))
	}
	cf := proto.CommunityFile{Founders: fpubs, Chain: []proto.ModSet{s}}
	if err := (&proto.Registry{}).Add(cf); err != nil { // refuse to emit an invalid file
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(cf)
}

// netFlags are shared by mod and client.
type netFlags struct {
	relays, bootstrap, listen, communities, trace, key *string
	private                                            *bool
}

func addNetFlags(fs *flag.FlagSet, private bool) netFlags {
	return netFlags{
		relays:      fs.String("relays", os.Getenv("LIBR_RELAYS"), "relay multiaddrs to reserve on (with -private)"),
		bootstrap:   fs.String("bootstrap", os.Getenv("LIBR_BOOTSTRAP"), "DHT bootstrap multiaddrs (DB nodes)"),
		listen:      fs.String("listen", env("LIBR_LISTEN", "/ip4/0.0.0.0/tcp/0"), "listen multiaddrs"),
		communities: fs.String("communities", os.Getenv("LIBR_COMMUNITIES"), "directory of community *.json files"),
		trace:       fs.String("trace", "", "append one JSON line per network request to this file"),
		key:         fs.String("key", "", "Ed25519 seed file (created if missing; empty: fresh key)"),
		private:     fs.Bool("private", private, "behind NAT: reserve on relays, advertise circuit addrs"),
	}
}

type node struct {
	host host.Host
	dht  *dht.IpfsDHT
	reg  *proto.Registry
	key  ed25519.PrivateKey
}

func (nf netFlags) start(ctx context.Context) (*node, error) {
	if *nf.communities == "" {
		return nil, errors.New("-communities is required")
	}
	reg, err := proto.LoadRegistry(os.DirFS(*nf.communities))
	if err != nil {
		return nil, fmt.Errorf("communities: %w", err)
	}
	var key ed25519.PrivateKey
	if *nf.key != "" {
		if key, err = p2p.LoadOrCreateKey(*nf.key); err != nil {
			return nil, err
		}
	} else {
		_, key, _ = ed25519.GenerateKey(nil)
	}
	relays, err := addrInfos(*nf.relays)
	if err != nil {
		return nil, fmt.Errorf("-relays: %w", err)
	}
	boot, err := addrInfos(*nf.bootstrap)
	if err != nil {
		return nil, fmt.Errorf("-bootstrap: %w", err)
	}
	if len(boot) == 0 {
		return nil, errors.New("-bootstrap is required")
	}
	h, err := p2p.NewHost(p2p.Config{Key: key, Listen: split(*nf.listen), Relays: relays, Private: *nf.private})
	if err != nil {
		return nil, err
	}
	d, err := p2p.NewDHT(ctx, h, false, boot)
	if err != nil {
		h.Close()
		return nil, err
	}
	n := &node{host: h, dht: d, reg: reg, key: key}
	jctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := p2p.Join(jctx, h, d, boot); err != nil {
		n.close()
		return nil, err
	}
	return n, nil
}

func (n *node) close() {
	n.dht.Close()
	n.host.Close()
}

func modServe(args []string) error {
	fs := flag.NewFlagSet("mod serve", flag.ContinueOnError)
	nf := addNetFlags(fs, true)
	policy := fs.String("policy", "approve-all", "approve-all | keyword:w1,w2 (deterministic; no LLM)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *nf.key == "" {
		return errors.New("mod serve: -key is required (the moderator's identity)")
	}
	var pol moderator.Policy
	switch {
	case *policy == "approve-all":
		pol = moderator.ApproveAll
	case strings.HasPrefix(*policy, "keyword:"):
		pol = moderator.Keywords(split(strings.TrimPrefix(*policy, "keyword:"))...)
	default:
		return fmt.Errorf("unknown policy %q", *policy)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	n, err := nf.start(ctx)
	if err != nil {
		return err
	}
	defer n.close()
	(&moderator.Mod{Key: n.key, Sets: n.reg, Policy: pol}).Register(n.host)
	go p2p.KeepFindable(ctx, n.host, n.dht, 30*time.Second)
	log.Printf("mod %s serving %v; policy %s", n.host.ID(), n.host.Addrs(), *policy)
	<-ctx.Done()
	return nil
}

func clientCmd(cmd string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	nf := addNetFlags(fs, false)
	community := fs.String("community", "", "community name")
	text := fs.String("text", "", "submit: post text")
	since := fs.Duration("since", time.Hour, "feed: how far back")
	asJSON := fs.Bool("json", false, "feed: print certificates as JSON lines")
	timeout := fs.Duration("timeout", 2*time.Minute, "overall timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	n, err := nf.start(ctx)
	if err != nil {
		return err
	}
	defer n.close()
	c := &forum.Client{Host: n.host, DHT: n.dht, Sets: n.reg, Key: n.key}
	if *nf.trace != "" {
		f, err := os.OpenFile(*nf.trace, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		var mu sync.Mutex
		enc := json.NewEncoder(f)
		c.Trace = func(e forum.Event) {
			mu.Lock()
			defer mu.Unlock()
			_ = enc.Encode(e)
		}
	}

	if cmd == "client submit" {
		start := time.Now()
		p, err := c.Submit(ctx, *community, []byte(*text))
		if err != nil {
			return err
		}
		id := p.Cert.Msg.ID()
		fmt.Fprintf(out, "posted hash=%s msg_id=%s votes=%d receipts=%d indexed=%d in %v\n",
			hex.EncodeToString(p.Hash[:]), hex.EncodeToString(id[:]), len(p.Cert.Votes), len(p.Receipts), len(p.Indexed),
			time.Since(start).Round(time.Millisecond))
		select { // let replication to the remaining replicas finish before exiting
		case <-p.Done:
		case <-ctx.Done():
		}
		return nil
	}

	now := time.Now()
	certs, err := c.Feed(ctx, *community, now.Add(-*since), now)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	for _, cert := range certs {
		if *asJSON {
			if err := enc.Encode(cert); err != nil {
				return err
			}
			continue
		}
		author, _ := proto.PeerIDFor(cert.Msg.Author)
		fmt.Fprintf(out, "%s\t%s\t%s\n", time.Unix(cert.Msg.Ts, 0).UTC().Format(time.RFC3339), shortID(author), cert.Msg.Content)
	}
	return nil
}

func shortID(id string) string {
	if p, err := peer.Decode(id); err == nil {
		return p.ShortString()
	}
	return id
}

func readKey(path string) (ed25519.PrivateKey, error) {
	seed, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("%s: want a %d-byte seed", path, ed25519.SeedSize)
	}
	return ed25519.NewKeyFromSeed(seed), nil
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
