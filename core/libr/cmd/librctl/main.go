// Command librctl is the headless LIBR v2 tool: key and genesis generation, a
// moderator daemon, and a posting/reading client. Used by the testbed and for
// measurements (--trace writes one JSON line per network request).
//
//	librctl keygen  -out FILE [-libp2p]
//	librctl genesis -community NAME -f 1 -mods PUBKEY,... (-founders KEYFILE,... | -founder-keys PUBKEY,... -out FILE)
//	librctl sign    -key FOUNDERKEY -in FILE     (each founder, on their own machine)
//	librctl verify  -in FILE
//	librctl mod serve -key FILE -communities DIR -policy approve-all|keyword:w1,w2 -votes FILE [-queue FILE -review w1,w2] [net flags]
//	librctl mod queue [-admin ADDR] list | approve KEY | reject KEY
//	librctl client submit -communities DIR -community NAME -text TEXT [net flags]
//	librctl client feed   -communities DIR -community NAME [-since 1h] [-json] [net flags]
//	librctl client report -communities DIR -hash HEX -reason TEXT [net flags]
//	librctl client delete -communities DIR -hash HEX -key AUTHORKEY [net flags]
//	librctl client bench  -communities DIR -community NAME [-n 20] [-reads 10] [-index-readers K] -trace FILE [net flags]
//	librctl client check  -communities DIR -in BENCH_OUTPUT [net flags]
//
// Net flags: -relays, -bootstrap (comma-separated multiaddrs with /p2p/ID), -listen,
// -communities, -seeds URLS, -cache FILE, -private, -trace FILE. Bootstrap peers come
// from -bootstrap, the shipped seed list, -cache, then -seeds servers. The first four default to the DB node's
// environment variables LIBR_RELAYS, LIBR_BOOTSTRAP, LIBR_LISTEN, LIBR_COMMUNITIES.
package main

import (
	"bytes"
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
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
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
	"github.com/libr-forum/Libr/core/libr/seeds"
)

func main() {
	log.SetFlags(log.Ltime)
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "librctl:", err)
		os.Exit(1)
	}
}

const usage = "usage: librctl keygen | genesis | mod serve | mod queue | client submit | client feed | client report | client delete | client bench | client check (see -h of each)"

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
	case "sign":
		return signGenesis(args, out)
	case "verify":
		return verifyGenesis(args, out)
	case "mod serve":
		return modServe(args)
	case "mod queue":
		return modQueue(args, out)
	case "client submit", "client feed", "client report", "client delete", "client bench", "client check":
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
	f := fs.Uint64("f", 1, "faulty moderators tolerated (needs exactly 3f+1 mods)")
	founders := fs.String("founders", "", "comma-separated founder seed files (all keys on this machine: dev only)")
	founderKeys := fs.String("founder-keys", "", "comma-separated founder public keys (base64): write an unsigned file for `sign`")
	outPath := fs.String("out", "", "with -founder-keys: file to write")
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
	if *founderKeys != "" {
		if *founders != "" || *outPath == "" {
			return errors.New("genesis: -founder-keys needs -out and excludes -founders")
		}
		cf := proto.CommunityFile{Chain: []proto.ModSet{s}}
		for _, b64 := range split(*founderKeys) {
			pk, err := base64.StdEncoding.DecodeString(b64)
			if err != nil || len(pk) != ed25519.PublicKeySize {
				return fmt.Errorf("founder key %q: not a base64 Ed25519 public key", b64)
			}
			cf.Founders = append(cf.Founders, pk)
		}
		if err := writeCommunity(*outPath, cf); err != nil {
			return err
		}
		fmt.Fprintf(out, "wrote %s: unsigned; each founder runs `librctl sign -key <their key> -in %s`\n", *outPath, *outPath)
		return nil
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

// signGenesis adds one founder's signature to an unsigned or partly signed genesis.
func signGenesis(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	keyPath := fs.String("key", "", "founder seed file")
	in := fs.String("in", "", "community file to sign in place")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cf, err := readCommunity(*in)
	if err != nil {
		return err
	}
	k, err := readKey(*keyPath)
	if err != nil {
		return err
	}
	pub := k.Public().(ed25519.PublicKey)
	listed := false
	for _, f := range cf.Founders {
		listed = listed || bytes.Equal(f, pub)
	}
	if !listed {
		return errors.New("sign: this key is not one of the file's founders")
	}
	g := &cf.Chain[0]
	for _, sg := range g.Sigs {
		if bytes.Equal(sg.Signer, pub) {
			return errors.New("sign: already signed by this key")
		}
	}
	proto.SignModSet(k, g)
	if err := writeCommunity(*in, cf); err != nil {
		return err
	}
	fmt.Fprintf(out, "signed: %d of %d founders\n", len(g.Sigs), len(cf.Founders))
	return nil
}

// verifyGenesis checks a community file exactly as nodes and apps will.
func verifyGenesis(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	in := fs.String("in", "", "community file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cf, err := readCommunity(*in)
	if err != nil {
		return err
	}
	if err := (&proto.Registry{}).Add(cf); err != nil {
		return err
	}
	g := cf.Chain[0]
	fmt.Fprintf(out, "ok: community %q, %d moderators, threshold %d (f=%d), signed by all %d founders\n",
		g.Community, len(g.Mods), g.Threshold, g.F, len(cf.Founders))
	return nil
}

func readCommunity(path string) (proto.CommunityFile, error) {
	var cf proto.CommunityFile
	b, err := os.ReadFile(path)
	if err != nil {
		return cf, err
	}
	if err := json.Unmarshal(b, &cf); err != nil {
		return cf, fmt.Errorf("%s: %w", path, err)
	}
	if len(cf.Chain) != 1 {
		return cf, fmt.Errorf("%s: want a genesis file (one moderator set), got %d", path, len(cf.Chain))
	}
	return cf, nil
}

func writeCommunity(path string, cf proto.CommunityFile) error {
	b, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// netFlags are shared by mod and client.
type netFlags struct {
	relays, bootstrap, seedURLs, cache, listen, communities, trace, key *string
	private                                                             *bool
}

func addNetFlags(fs *flag.FlagSet, private bool) netFlags {
	return netFlags{
		relays:      fs.String("relays", os.Getenv("LIBR_RELAYS"), "relay multiaddrs to reserve on (with -private)"),
		bootstrap:   fs.String("bootstrap", os.Getenv("LIBR_BOOTSTRAP"), "DHT bootstrap multiaddrs (DB nodes)"),
		seedURLs:    fs.String("seeds", os.Getenv("LIBR_SEED_URLS"), "seed servers (librserver) to fetch bootstrap peers from"),
		cache:       fs.String("cache", "", "peer cache file: read for bootstrap, rewritten after joining"),
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
	boot = seeds.Resolve(ctx, boot, *nf.cache, split(*nf.seedURLs), 8)
	if len(boot) == 0 {
		return nil, errors.New("no bootstrap peers: set -bootstrap, -seeds or -cache")
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
	if *nf.cache != "" {
		if err := seeds.SaveCache(*nf.cache, h, d); err != nil {
			log.Printf("peer cache: %v", err)
		}
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
	policy := fs.String("policy", "approve-all", "posts: approve-all | keyword:w1,w2 | llm:FILE (local Ollama model, rules per community; see llm.go)")
	reportPolicy := fs.String("report-policy", "", "reports: remove what this policy would reject (default: -policy); llm:FILE also weighs the reporter's reason")
	voteLog := fs.String("votes", "", "file of every decision this moderator signs (required; a restart must never sign twice). \"memory\": keep in memory only (tests)")
	review := fs.String("review", "", "posts containing any of these comma-separated words, or images, go to the human queue (needs -queue)")
	queuePath := fs.String("queue", "", "human review queue file; with it, every report goes to a human unless -report-policy is set")
	admin := fs.String("admin", "127.0.0.1:7070", "with -queue: loopback address for `librctl mod queue` (list and decide)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *nf.key == "" {
		return errors.New("mod serve: -key is required (the moderator's identity)")
	}
	pol, judgesImages, err := parsePolicy(*policy)
	if err != nil {
		return err
	}
	switch *voteLog {
	case "":
		return errors.New("mod serve: -votes FILE is required (use -votes memory only for throwaway tests)")
	case "memory":
		*voteLog = ""
	}
	mod := &moderator.Mod{Policy: pol, VoteLog: *voteLog}
	if *reportPolicy != "" {
		if strings.HasPrefix(*reportPolicy, "llm:") {
			if mod.Report, err = loadLLMReport(strings.TrimPrefix(*reportPolicy, "llm:")); err != nil {
				return err
			}
		} else {
			rp, _, err := parsePolicy(*reportPolicy)
			if err != nil {
				return err
			}
			mod.Report = moderator.RemoveIfRejected(rp)
		}
	}
	if *queuePath != "" {
		if mod.Queue, err = moderator.NewQueue(*queuePath); err != nil {
			return err
		}
		mod.Policy = withReview(pol, split(*review), !judgesImages)
		if mod.Report == nil {
			mod.Report = func(context.Context, proto.Msg, string) (bool, error) { return false, moderator.ErrDefer }
		}
		if err := serveAdmin(*admin, mod); err != nil {
			return err
		}
	} else if *review != "" {
		return errors.New("mod serve: -review needs -queue")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	n, err := nf.start(ctx)
	if err != nil {
		return err
	}
	defer n.close()
	mod.Key, mod.Sets = n.key, n.reg
	mod.Register(n.host)
	go p2p.KeepFindable(ctx, n.host, n.dht, 30*time.Second)
	log.Printf("mod %s serving %v; policy %s", n.host.ID(), n.host.Addrs(), *policy)
	<-ctx.Done()
	return nil
}

// withReview defers to the human queue posts that p approves but that hold one of
// words, or an image unless p judges images itself; p's rejections stand.
func withReview(p moderator.Policy, words []string, images bool) moderator.Policy {
	if images {
		words = append(words, "data:image/")
	}
	if len(words) == 0 {
		return p
	}
	review := moderator.Keywords(words...)
	return func(ctx context.Context, m proto.Msg) (bool, error) {
		ok, err := p(ctx, m)
		if err != nil || !ok {
			return ok, err
		}
		if clean, _ := review(ctx, m); !clean {
			return false, moderator.ErrDefer
		}
		return true, nil
	}
}

// serveAdmin lists the queue (GET /queue) and records decisions
// (POST /decide?key=K&approve=true|false) on a loopback address only.
func serveAdmin(addr string, mod *moderator.Mod) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("-admin %s: must be a loopback address", addr)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /queue", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(mod.Queue.Items())
	})
	mux.HandleFunc("POST /decide", func(w http.ResponseWriter, r *http.Request) {
		if err := mod.Resolve(r.URL.Query().Get("key"), r.URL.Query().Get("approve") == "true"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "ok")
	})
	go func() { _ = http.Serve(ln, mux) }()
	return nil
}

// modQueue is the human side of a server moderator: list the queue, or approve or
// reject one item (for a report, approve means remove the post).
func modQueue(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("mod queue", flag.ContinueOnError)
	admin := fs.String("admin", "127.0.0.1:7070", "the -admin address of mod serve")
	if err := fs.Parse(args); err != nil {
		return err
	}
	base := "http://" + *admin
	var resp *http.Response
	var err error
	switch a := fs.Args(); {
	case len(a) == 0 || a[0] == "list":
		resp, err = http.Get(base + "/queue")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var items []moderator.Item
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			return err
		}
		for _, it := range items {
			state := "undecided"
			if it.Vote != nil {
				state = map[bool]string{true: "approved", false: "rejected"}[it.Vote.Approve]
			}
			fmt.Fprintf(out, "%s\t%s\t%s\t%q\t%q\n", it.Key, it.Kind, state, it.Reason, trim(string(it.Cert.Msg.Content), 80))
		}
		return nil
	case len(a) == 2 && (a[0] == "approve" || a[0] == "reject"):
		resp, err = http.Post(base+"/decide?key="+url.QueryEscape(a[1])+"&approve="+fmt.Sprint(a[0] == "approve"), "", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return errors.New(strings.TrimSpace(string(b)))
		}
		fmt.Fprint(out, string(b))
		return nil
	}
	return errors.New("usage: librctl mod queue [-admin ADDR] list | approve KEY | reject KEY")
}

func trim(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

// parsePolicy also reports whether the policy judges attached images itself.
func parsePolicy(spec string) (moderator.Policy, bool, error) {
	switch {
	case spec == "approve-all":
		return moderator.ApproveAll, false, nil
	case strings.HasPrefix(spec, "keyword:"):
		return moderator.Keywords(split(strings.TrimPrefix(spec, "keyword:"))...), false, nil
	case strings.HasPrefix(spec, "llm:"):
		return loadLLMPolicy(strings.TrimPrefix(spec, "llm:"))
	}
	return nil, false, fmt.Errorf("unknown policy %q", spec)
}

func clientCmd(cmd string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	nf := addNetFlags(fs, false)
	community := fs.String("community", "", "community name")
	text := fs.String("text", "", "submit: post text")
	hashHex := fs.String("hash", "", "report/delete: certificate hash (hex, as printed by submit)")
	reason := fs.String("reason", "", "report: why")
	since := fs.Duration("since", time.Hour, "feed: how far back")
	asJSON := fs.Bool("json", false, "feed: print certificates as JSON lines")
	timeout := fs.Duration("timeout", 2*time.Minute, "overall timeout")
	benchN := fs.Int("n", 20, "bench: posts to make, one after another")
	benchReads := fs.Int("reads", 10, "bench: feed reads after posting")
	readers := fs.Int("index-readers", 0, "index nodes read per feed bucket (0: default)")
	checkIn := fs.String("in", "", "check: bench output whose posts to fetch again by hash")
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
	c := &forum.Client{Host: n.host, DHT: n.dht, Sets: n.reg, Key: n.key, IndexReaders: *readers}
	var tag atomic.Value // bench: which operation trace events belong to
	tag.Store("")
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
			_ = enc.Encode(struct {
				Tag string `json:"tag,omitempty"`
				forum.Event
			}{tag.Load().(string), e})
		}
	}

	if cmd == "client bench" {
		return bench(ctx, c, *community, *benchN, *benchReads, func(t string) { tag.Store(t) }, out)
	}
	if cmd == "client check" {
		return check(ctx, c, *checkIn, out)
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

	if cmd == "client report" || cmd == "client delete" {
		var h proto.Hash
		if b, err := hex.DecodeString(*hashHex); err != nil || copy(h[:], b) != len(h) || len(b) != len(h) {
			return fmt.Errorf("-hash: want %d hex bytes", len(h))
		}
		cert, err := c.Get(ctx, h)
		if err != nil {
			return err
		}
		var r *forum.Removed
		if cmd == "client report" {
			r, err = c.Report(ctx, cert, *reason)
		} else {
			r, err = c.Delete(ctx, cert)
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "removed msg_id=%s kind=%s acks=%d index_acks=%d\n",
			hex.EncodeToString(r.Tomb.MsgID[:]), r.Tomb.Kind(), len(r.Acks), len(r.IndexAcks))
		select {
		case <-r.Done:
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

// bench posts n times, one after another (each waits for full replication, so trace
// events belong to one post), then reads the feed `reads` times. One JSON line per
// operation: its tag, outcome, end-to-end milliseconds and, for reads, posts seen.
func bench(ctx context.Context, c *forum.Client, community string, n, reads int, setTag func(string), out io.Writer) error {
	enc := json.NewEncoder(out)
	type result struct {
		Tag   string  `json:"tag"`
		Ms    float64 `json:"ms"`
		Err   string  `json:"err,omitempty"`
		Seen  int     `json:"seen,omitempty"`
		Votes int     `json:"votes,omitempty"`
		Bytes int     `json:"cert_bytes,omitempty"`
		Hash  string  `json:"hash,omitempty"`
	}
	emit := func(r result, start time.Time, err error) error {
		r.Ms = float64(time.Since(start).Microseconds()) / 1000
		if err != nil {
			r.Err = err.Error()
		}
		return enc.Encode(r)
	}
	since := time.Now().Add(-time.Minute)
	for i := 0; i < n && ctx.Err() == nil; i++ {
		tag := fmt.Sprintf("post-%d", i)
		setTag(tag)
		// A fresh author per post: moderators allow one author 0.2 posts/s.
		if _, k, err := ed25519.GenerateKey(nil); err == nil {
			c.Key = k
		}
		start := time.Now()
		p, err := c.Submit(ctx, community, []byte(fmt.Sprintf("bench %d %d", i, start.UnixNano())))
		r := result{Tag: tag}
		if err == nil {
			r.Votes, r.Hash = len(p.Cert.Votes), hex.EncodeToString(p.Hash[:])
			if b, jerr := json.Marshal(p.Cert); jerr == nil {
				r.Bytes = len(b)
			}
		}
		if err := emit(r, start, err); err != nil {
			return err
		}
		if err == nil {
			select {
			case <-p.Done:
			case <-ctx.Done():
			}
		}
		// Moderators allow one peer 5 requests/s: stay under it.
		time.Sleep(time.Until(start.Add(250 * time.Millisecond)))
	}
	for i := 0; i < reads && ctx.Err() == nil; i++ {
		tag := fmt.Sprintf("feed-%d", i)
		setTag(tag)
		start := time.Now()
		certs, err := c.Feed(ctx, community, since, time.Now())
		if err := emit(result{Tag: tag, Seen: len(certs)}, start, err); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// check fetches every post listed in a bench output by its hash and prints one JSON
// line per post: whether a verified copy came back, and how long it took.
func check(ctx context.Context, c *forum.Client, in string, out io.Writer) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	for _, line := range strings.Split(string(b), "\n") {
		var r struct {
			Hash string `json:"hash"`
		}
		var h proto.Hash
		if json.Unmarshal([]byte(line), &r) != nil || r.Hash == "" {
			continue
		}
		if raw, err := hex.DecodeString(r.Hash); err != nil || copy(h[:], raw) != len(h) {
			continue
		}
		start := time.Now()
		_, err := c.Get(ctx, h)
		res := struct {
			Hash string  `json:"hash"`
			OK   bool    `json:"ok"`
			Ms   float64 `json:"ms"`
			Err  string  `json:"err,omitempty"`
		}{Hash: r.Hash, OK: err == nil, Ms: float64(time.Since(start).Microseconds()) / 1000}
		if err != nil {
			res.Err = err.Error()
		}
		if err := enc.Encode(res); err != nil {
			return err
		}
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
