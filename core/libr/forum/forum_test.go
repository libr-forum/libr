package forum

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	kb "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/dbnode"
	"github.com/libr-forum/Libr/core/libr/internal/testnet"
	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
)

func key(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("forum-test/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

func pub(k ed25519.PrivateKey) ed25519.PublicKey { return k.Public().(ed25519.PublicKey) }

// registry: community "general", 4 mods, f=1, threshold 3.
func registry(t *testing.T) (*proto.Registry, []ed25519.PrivateKey) {
	t.Helper()
	s := proto.ModSet{Community: "general", F: 1, Threshold: 3}
	var mods []ed25519.PrivateKey
	for i := 0; i < 4; i++ {
		k := key(fmt.Sprint("mod", i))
		id, _ := proto.PeerIDFor(pub(k))
		mods = append(mods, k)
		s.Mods = append(s.Mods, proto.ModEntry{PubKey: pub(k), PeerID: id})
	}
	proto.SignModSet(key("founder"), &s)
	reg := &proto.Registry{}
	if err := reg.Add(proto.CommunityFile{Founders: []ed25519.PublicKey{pub(key("founder"))}, Chain: []proto.ModSet{s}}); err != nil {
		t.Fatal(err)
	}
	return reg, mods
}

// tamperStore serves altered certificates once Evil is set (a malicious replica).
type tamperStore struct {
	*dbnode.MemStore
	mu   sync.Mutex
	Evil bool
}

func (s *tamperStore) Get(h proto.Hash) (*proto.MsgCert, error) {
	c, err := s.MemStore.Get(h)
	s.mu.Lock()
	evil := s.Evil
	s.mu.Unlock()
	if c != nil && evil {
		forged := *c
		forged.Msg.Content = []byte("forged by a DB node")
		return &forged, nil
	}
	return c, err
}

type world struct {
	t       *testing.T
	reg     *proto.Registry
	modKeys []ed25519.PrivateKey
	relays  []host.Host
	dbs     []*testnet.Node
	stores  []*tamperStore
	mods    []*testnet.Node
	client  *Client
	trace   []Event
	mu      sync.Mutex
}

// newWorld starts relays and NATed DB nodes, moderators and one author, all reachable
// only through circuits, with the DHT joined and moderators announced.
func newWorld(t *testing.T, nRelays, nDBs int, policy moderator.Policy, report ...moderator.ReportPolicy) *world {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	w := &world{t: t}
	reg, modKeys := registry(t)
	w.reg, w.modKeys = reg, modKeys
	var relays []peer.AddrInfo
	for i := 0; i < nRelays; i++ {
		r, err := testnet.NewRelay("/ip4/127.0.0.1/tcp/0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { r.Close() })
		w.relays = append(w.relays, r)
		relays = append(relays, testnet.AddrInfo(r))
	}
	start := func(k ed25519.PrivateKey, server bool) *testnet.Node {
		n, err := testnet.NewNATedHost(ctx, p2p.Config{Key: k, Listen: []string{"/ip4/127.0.0.1/tcp/0"}}, relays, server)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { n.Close() })
		return n
	}
	for i := 0; i < nDBs; i++ {
		k := key(fmt.Sprint("db", i))
		n := start(k, true)
		st := &tamperStore{MemStore: dbnode.NewMemStore()}
		(&dbnode.Node{Key: k, Sets: reg, Store: st}).Register(n.Host)
		w.dbs, w.stores = append(w.dbs, n), append(w.stores, st)
	}
	for _, k := range modKeys {
		n := start(k, false)
		m := &moderator.Mod{Key: k, Sets: reg, Policy: policy}
		if len(report) > 0 {
			m.Report = report[0]
		}
		m.Register(n.Host)
		w.mods = append(w.mods, n)
	}
	cl := start(nil, false)
	all := append(append(append([]*testnet.Node{}, w.dbs...), w.mods...), cl)
	for _, n := range all[1:] {
		if err := n.Host.Connect(ctx, testnet.AddrInfo(w.dbs[0].Host)); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range all {
		<-n.DHT.RefreshRoutingTable()
	}
	for _, m := range w.mods {
		if err := testnet.Announce(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	w.client = &Client{Host: cl.Host, DHT: cl.DHT, Sets: reg, Key: key("author"), Trace: func(e Event) {
		w.mu.Lock()
		w.trace = append(w.trace, e)
		w.mu.Unlock()
	}}
	return w
}

func (w *world) submit(text string) (*Posted, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	return w.client.Submit(ctx, "general", []byte(text))
}

func (w *world) feed() []string {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	now := time.Now()
	certs, err := w.client.Feed(ctx, "general", now.Add(-time.Hour), now)
	if err != nil {
		w.t.Fatalf("feed: %v", err)
	}
	var out []string
	for _, c := range certs {
		out = append(out, string(c.Msg.Content))
	}
	return out
}

func (w *world) dbIndex(id peer.ID) int {
	for i, d := range w.dbs {
		if d.Host.ID() == id {
			return i
		}
	}
	return -1
}

func TestSubmitAndFeed(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 12 hosts")
	}
	w := newWorld(t, 1, 6, moderator.ApproveAll)

	p1, err := w.submit("first post")
	if err != nil {
		t.Fatal(err)
	}
	if len(p1.Receipts) < StoredQuorum || len(p1.Indexed) < StoredQuorum || len(p1.Cert.Votes) < 3 {
		t.Fatalf("receipts %d, indexed %d, votes %d", len(p1.Receipts), len(p1.Indexed), len(p1.Cert.Votes))
	}
	<-p1.Done
	held := 0
	for _, s := range w.stores {
		if c, _ := s.Get(p1.Hash); c != nil {
			held++
		}
	}
	if held != Replicas {
		t.Errorf("cert held by %d DB nodes after background replication, want %d", held, Replicas)
	}
	time.Sleep(1100 * time.Millisecond) // distinct timestamps for a stable order
	if _, err := w.submit("second post"); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(w.feed(), "|"); got != "second post|first post" {
		t.Fatalf("feed = %q", got)
	}
	ops := map[string]int{}
	for _, e := range w.trace {
		if e.Err == "" && e.Bytes > 0 {
			ops[e.Op]++
		}
	}
	for _, op := range []string{"vote", "store", "index_put", "index_get", "get"} {
		if ops[op] == 0 {
			t.Errorf("no successful %q in trace", op)
		}
	}
}

// TestTamperedReplicaSkipped: the nearest replica serves a forged copy; the reader
// rejects it (hash mismatch) and takes the next replica's genuine copy.
func TestTamperedReplicaSkipped(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 12 hosts")
	}
	w := newWorld(t, 1, 6, moderator.ApproveAll)
	p, err := w.submit("genuine")
	if err != nil {
		t.Fatal(err)
	}
	<-p.Done
	ids := make([]peer.ID, len(w.dbs))
	for i, d := range w.dbs {
		ids[i] = d.Host.ID()
	}
	nearest := kb.SortClosestPeers(ids, kb.ConvertKey(string(p.Hash[:])))[0]
	evil := w.stores[w.dbIndex(nearest)]
	evil.mu.Lock()
	evil.Evil = true
	evil.mu.Unlock()

	if got := w.feed(); len(got) != 1 || got[0] != "genuine" {
		t.Fatalf("feed = %q", got)
	}
	rejected := false
	for _, e := range w.trace {
		if e.Op == "get" && e.Peer == nearest.String() && strings.Contains(e.Err, "hash") {
			rejected = true
		}
	}
	if !rejected {
		t.Error("forged copy from the nearest replica was never seen and rejected")
	}
}

func TestModeratorFaults(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	w := newWorld(t, 1, 4, moderator.Keywords("spam"))

	if _, err := w.submit("buy spam now"); !errors.Is(err, ErrRejected) {
		t.Fatalf("policy rejection: err = %v, want ErrRejected", err)
	}
	w.mods[0].Close()
	if _, err := w.submit("one mod down (f=1)"); err != nil {
		t.Fatalf("one of four mods down: %v", err)
	}
	w.mods[1].Close()
	start := time.Now()
	if _, err := w.submit("two mods down"); !errors.Is(err, ErrNoQuorum) {
		t.Fatalf("two of four mods down: err = %v, want ErrNoQuorum", err)
	}
	t.Logf("no-quorum detected in %v", time.Since(start).Round(time.Millisecond))
}

// TestDBNodeDown: with K=4 and only 4 DB nodes, one dead node leaves exactly the
// receipt quorum of 3.
func TestDBNodeDown(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	w := newWorld(t, 1, 4, moderator.ApproveAll)
	w.dbs[3].Close()
	p, err := w.submit("one DB node down")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Receipts) != StoredQuorum {
		t.Fatalf("receipts %d, want %d", len(p.Receipts), StoredQuorum)
	}
	if got := w.feed(); len(got) != 1 {
		t.Fatalf("feed = %q", got)
	}
}

// TestRelayFailover: every peer holds reservations on two relays; losing one leaves
// posting and reading working through the other.
func TestRelayFailover(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 13 hosts")
	}
	w := newWorld(t, 2, 4, moderator.ApproveAll)
	if _, err := w.submit("before"); err != nil {
		t.Fatal(err)
	}
	w.relays[0].Close()
	if _, err := w.submit("after relay loss"); err != nil {
		t.Fatalf("submit after relay loss: %v", err)
	}
	if got := w.feed(); len(got) != 2 {
		t.Fatalf("feed after relay loss = %q", got)
	}
}
