package dbnode

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/libr-forum/Libr/core/libr/internal/testnet"
	"github.com/libr-forum/Libr/core/libr/p2p"
	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

func key(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("dbnode-test/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

func pub(k ed25519.PrivateKey) ed25519.PublicKey { return k.Public().(ed25519.PublicKey) }

// community registers "general" with 4 mods (f=1, threshold 3).
func community(t *testing.T) (*proto.Registry, []ed25519.PrivateKey) {
	t.Helper()
	s := proto.ModSet{Community: "general", F: 1, Threshold: 3}
	var mods []ed25519.PrivateKey
	for i := 0; i < 4; i++ {
		k := key(fmt.Sprint("mod", i))
		id, err := proto.PeerIDFor(pub(k))
		if err != nil {
			t.Fatal(err)
		}
		mods = append(mods, k)
		s.Mods = append(s.Mods, proto.ModEntry{PubKey: pub(k), PeerID: id})
	}
	founder := key("founder")
	proto.SignModSet(founder, &s)
	reg := &proto.Registry{}
	if err := reg.Add(proto.CommunityFile{Founders: []ed25519.PublicKey{pub(founder)}, Chain: []proto.ModSet{s}}); err != nil {
		t.Fatal(err)
	}
	return reg, mods
}

func cert(author ed25519.PrivateKey, content string, ts int64, approvers ...ed25519.PrivateKey) proto.MsgCert {
	c := proto.SignMsg(author, proto.Msg{Community: "general", Ts: ts, Content: []byte(content)})
	for _, m := range approvers {
		c.Votes = append(c.Votes, proto.SignModVote(m, c.Msg, 0, true))
	}
	return c
}

type rig struct {
	client *testnet.Node
	db     peer.ID
	dbKey  ed25519.PrivateKey
	store  *MemStore
}

// newRig starts a relay, a NATed DB node and a NATed client; the client reaches the DB
// node only through a relayed circuit.
func newRig(t *testing.T, reg *proto.Registry) *rig {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	relay, err := testnet.NewRelay("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { relay.Close() })
	relays := []peer.AddrInfo{testnet.AddrInfo(relay)}

	dbKey := key("db")
	db, err := testnet.NewNATedHost(ctx, p2p.Config{Key: dbKey, Listen: []string{"/ip4/127.0.0.1/tcp/0"}}, relays, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := NewMemStore()
	(&Node{Key: dbKey, Sets: reg, Store: store}).Register(db.Host)

	cl, err := testnet.NewNATedHost(ctx, p2p.Config{Listen: []string{"/ip4/127.0.0.1/tcp/0"}}, relays, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cl.Close() })
	if err := cl.Host.Connect(ctx, testnet.AddrInfo(db.Host)); err != nil {
		t.Fatal(err)
	}
	return &rig{client: cl, db: db.Host.ID(), dbKey: dbKey, store: store}
}

func (r *rig) call(t *testing.T, p string, req, resp any) {
	t.Helper()
	if err := wire.Call(context.Background(), r.client.Host, r.db, protocol.ID(p), req, resp); err != nil {
		t.Fatalf("%s: %v", p, err)
	}
}

var now = time.Now().Unix()

func TestStoreGetIndexOverRelay(t *testing.T) {
	reg, mods := community(t)
	r := newRig(t, reg)
	author := key("author")
	good := cert(author, "hello v2", now, mods[0], mods[1], mods[2])

	var rr wire.ReceiptResp
	r.call(t, string(wire.ProtoStore), good, &rr)
	if rr.Err != "" || rr.Receipt == nil {
		t.Fatalf("store refused valid cert: %+v", rr)
	}
	rc := *rr.Receipt
	if !rc.Verify() || rc.CertHash != good.Hash() || rc.Community != "general" {
		t.Fatalf("bad receipt: %+v", rc)
	}
	// A client can tie the receipt to the peer it talked to: the node's peer ID derives
	// from the key that signed the receipt.
	if id, _ := proto.PeerIDFor(rc.Node); id != r.db.String() {
		t.Fatalf("receipt key %s does not match node peer %s", id, r.db)
	}

	var gr wire.GetResp
	r.call(t, string(wire.ProtoGet), wire.GetReq{Hash: good.Hash()}, &gr)
	if gr.Cert == nil || gr.Cert.Hash() != good.Hash() || proto.VerifyMsgCert(reg, *gr.Cert, time.Now()) != nil {
		t.Fatalf("get returned %+v", gr)
	}
	r.call(t, string(wire.ProtoGet), wire.GetReq{Hash: proto.Sum([]byte("nope"))}, &gr)
	if gr.Err != wire.ErrNotFound {
		t.Fatalf("missing cert: %+v", gr)
	}

	r.call(t, string(wire.ProtoIndexPut), good, &rr)
	if rr.Err != "" {
		t.Fatalf("index put: %+v", rr)
	}
	var ir wire.IndexGetResp
	r.call(t, string(wire.ProtoIndexGet), wire.IndexGetReq{Community: "general", Bucket: wire.Bucket(now)}, &ir)
	if len(ir.Hashes) != 1 || ir.Hashes[0] != good.Hash() {
		t.Fatalf("index: %+v", ir)
	}
}

func TestStoreRefusesInvalid(t *testing.T) {
	reg, mods := community(t)
	r := newRig(t, reg)
	author, outsider := key("author"), key("outsider")

	dup := cert(author, "dup", now, mods[0], mods[1])
	dup.Votes = append(dup.Votes, dup.Votes[0])
	unknown := cert(author, "x", now, mods[0], mods[1], mods[2])
	unknown.Msg.Community = "nope"

	cases := map[string]proto.MsgCert{
		"self-approved by outsider": cert(author, "spam", now, outsider),
		"below threshold":           cert(author, "few", now, mods[0], mods[1]),
		"duplicate votes":           dup,
		"unknown community":         unknown,
		"future timestamp":          cert(author, "later", now+3600, mods[0], mods[1], mods[2]),
	}
	for name, c := range cases {
		for _, p := range []string{string(wire.ProtoStore), string(wire.ProtoIndexPut)} {
			var rr wire.ReceiptResp
			r.call(t, p, c, &rr)
			if rr.Receipt != nil || !strings.HasPrefix(rr.Err, "invalid") {
				t.Errorf("%s via %s: %+v", name, p, rr)
			}
		}
		if got, _ := r.store.Get(c.Hash()); got != nil {
			t.Errorf("%s: stored anyway", name)
		}
	}
	if hs, _ := r.store.Index("general", wire.Bucket(now), 10); len(hs) != 0 {
		t.Errorf("invalid certs indexed: %v", hs)
	}
}

func TestOversizeAndRateLimit(t *testing.T) {
	reg, mods := community(t)
	r := newRig(t, reg)

	// A frame header claiming more than MaxFrame is refused before the body is read.
	s, err := p2p.LimitedOK{Host: r.client.Host}.NewStream(context.Background(), r.db, wire.ProtoStore)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.Write(binary.AppendUvarint(nil, wire.MaxFrame+1))
	_ = s.CloseWrite()
	var rr wire.ReceiptResp
	if err := wire.ReadFrame(bufio.NewReader(s), &rr); err != nil || !strings.Contains(rr.Err, "too large") {
		t.Fatalf("oversize: %+v %v", rr, err)
	}
	_ = s.Close()

	c := cert(key("author"), "flood", now, mods[0], mods[1], mods[2])
	limited := 0
	for i := 0; i < rateBurst+20; i++ {
		var rr wire.ReceiptResp
		r.call(t, string(wire.ProtoStore), c, &rr)
		if rr.Err == wire.ErrRateLimited.Error() {
			limited++
		}
	}
	if limited == 0 {
		t.Fatal("no request was rate limited")
	}
}
