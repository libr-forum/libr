package app

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/dbnode"
	"github.com/libr-forum/Libr/core/libr/internal/testnet"
	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
)

func key(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("app-test/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

type net struct {
	t        *testing.T
	relays   []string
	boot     []string
	comm     proto.CommunityFile
	modKeys  []ed25519.PrivateKey
	modDirs  []string
	mods     []*App
	finalMu  sync.Mutex
	finalled []PendingInfo
}

// newNet starts a relay and 4 NATed DB nodes, and the genesis of "general" whose 4
// moderators are App instances started by startMod.
func newNet(t *testing.T) *net {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	relay, err := testnet.NewRelay("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { relay.Close() })
	ri := testnet.AddrInfo(relay)
	n := &net{t: t, relays: []string{fmt.Sprintf("%s/p2p/%s", ri.Addrs[0], ri.ID)}}

	s := proto.ModSet{Community: "general", F: 1, Threshold: 3}
	for i := 0; i < 4; i++ {
		k := key(fmt.Sprint("mod", i))
		id, _ := proto.PeerIDFor(k.Public().(ed25519.PublicKey))
		s.Mods = append(s.Mods, proto.ModEntry{PubKey: k.Public().(ed25519.PublicKey), PeerID: id})
		n.modKeys = append(n.modKeys, k)
		n.modDirs = append(n.modDirs, t.TempDir())
	}
	proto.SignModSet(key("founder"), &s)
	n.comm = proto.CommunityFile{Founders: []ed25519.PublicKey{key("founder").Public().(ed25519.PublicKey)}, Chain: []proto.ModSet{s}}
	reg := &proto.Registry{}
	if err := reg.Add(n.comm); err != nil {
		t.Fatal(err)
	}

	var dbs []*testnet.Node
	for i := 0; i < 4; i++ {
		k := key(fmt.Sprint("db", i))
		d, err := testnet.NewNATedHost(ctx, p2p.Config{Key: k, Listen: []string{"/ip4/127.0.0.1/tcp/0"}}, []peer.AddrInfo{ri}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { d.Close() })
		(&dbnode.Node{Key: k, Sets: reg, Store: dbnode.NewMemStore()}).Register(d.Host)
		dbs = append(dbs, d)
	}
	for _, d := range dbs[1:] {
		if err := d.Host.Connect(ctx, testnet.AddrInfo(dbs[0].Host)); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range dbs {
		<-d.DHT.RefreshRoutingTable()
	}
	for _, a := range dbs[0].Host.Addrs() {
		n.boot = append(n.boot, fmt.Sprintf("%s/p2p/%s", a, dbs[0].Host.ID()))
	}
	for i := range n.modKeys {
		n.startMod(i)
	}
	return n
}

func (n *net) start(k ed25519.PrivateKey, dir string, policy moderator.Policy) *App {
	n.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	a, err := Start(ctx, Config{DataDir: dir, Key: k, Relays: n.relays, Bootstrap: n.boot, Policy: policy,
		Communities: []proto.CommunityFile{n.comm},
		OnFinalized: func(p PendingInfo) { n.finalMu.Lock(); n.finalled = append(n.finalled, p); n.finalMu.Unlock() }})
	if err != nil {
		n.t.Fatal(err)
	}
	n.t.Cleanup(a.Close)
	return a
}

// Moderators approve text and queue anything with an image for a human.
func imagesToHumans(_ context.Context, m proto.Msg) (bool, error) {
	if strings.Contains(string(m.Content), "<img") {
		return false, moderator.ErrDefer
	}
	return true, nil
}

func (n *net) startMod(i int) {
	a := n.start(n.modKeys[i], n.modDirs[i], imagesToHumans)
	if len(n.mods) <= i {
		n.mods = append(n.mods, a)
	} else {
		n.mods[i] = a
	}
}

// decide makes the first `count` moderators decide every queued item.
func (n *net) decide(count int, approve bool) {
	n.t.Helper()
	for _, m := range n.mods[:count] {
		for _, it := range m.ModQueue() {
			if it.Vote == nil {
				if err := m.ModDecide(it.Key, approve); err != nil {
					n.t.Fatal(err)
				}
			}
		}
	}
}

func feed(t *testing.T, a *App) []Post {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ps, err := a.Feed(ctx, "general", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return ps
}

func TestApp(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	n := newNet(t)
	author := n.start(key("author"), t.TempDir(), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if c := author.Communities(); len(c) != 1 || c[0].Name != "general" || c[0].IsMod || c[0].Threshold != 3 {
		t.Fatalf("author communities %+v", c)
	}
	if !n.mods[0].IsMod("general") || author.IsMod("general") {
		t.Fatal("IsMod wrong")
	}

	text, err := author.Post(ctx, "general", "hello v2")
	if err != nil || text.Status != "posted" {
		t.Fatalf("text post: %+v %v", text, err)
	}
	img, err := author.Post(ctx, "general", `<img src="data:image/png;base64,AAAA">`)
	if err != nil || img.Status != "pending" {
		t.Fatalf("image post: %+v %v", img, err)
	}
	if p := author.Pending(); len(p) != 1 || p[0].Waiting != 4 || p[0].Kind != "post" {
		t.Fatalf("pending %+v", p)
	}
	if q := n.mods[0].ModQueue(); len(q) != 1 || q[0].Kind != "post" {
		t.Fatalf("mod queue %+v", q)
	}
	n.decide(3, true)
	author.Tick(ctx)
	if p := author.Pending(); len(p) != 0 {
		t.Fatalf("still pending after review: %+v", p)
	}
	ps := feed(t, author)
	if len(ps) != 2 || !ps[0].Mine {
		t.Fatalf("feed %+v", ps)
	}

	// Report the text post: reports always go to humans; three remove it.
	rep, err := author.Report(ctx, text.Hash, "test removal")
	if err != nil || rep.Status != "pending" {
		t.Fatalf("report: %+v %v", rep, err)
	}
	n.decide(3, true)
	author.Tick(ctx)
	if ps := feed(t, author); len(ps) != 1 || strings.Contains(ps[0].Content, "hello") {
		t.Fatalf("feed after report %+v", ps)
	}

	// Author deletes the image post.
	if err := author.Delete(ctx, feed(t, author)[0].Hash); err != nil {
		t.Fatal(err)
	}
	if ps := feed(t, author); len(ps) != 0 {
		t.Fatalf("feed after delete %+v", ps)
	}

	// Two moderators offline: the post waits, then is re-signed and posted once they
	// are back (their restart keeps the same identity).
	n.mods[0].Close()
	n.mods[1].Close()
	off, err := author.Post(ctx, "general", "while mods are away")
	if err != nil || off.Status != "pending" {
		t.Fatalf("post with 2 mods down: %+v %v", off, err)
	}
	n.startMod(0)
	n.startMod(1)
	// The app retries every Tick; failed dials to the absent mods leave a short dial
	// backoff, so an immediate retry may still miss them.
	for start := time.Now(); len(author.Pending()) > 0; time.Sleep(time.Second) {
		if time.Since(start) > 45*time.Second {
			t.Fatalf("still pending: %+v", author.Pending())
		}
		author.Tick(ctx)
	}
	if ps := feed(t, author); len(ps) != 1 || ps[0].Content != "while mods are away" {
		t.Fatalf("feed after mods returned %+v", ps)
	}
	n.finalMu.Lock()
	defer n.finalMu.Unlock()
	var got []string
	for _, f := range n.finalled {
		got = append(got, f.Kind+":"+f.Status)
	}
	if strings.Join(got, ",") != "post:posted,report:removed,post:posted" {
		t.Fatalf("finalized events %v", got)
	}
}
