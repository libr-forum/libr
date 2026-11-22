package sqlstore

import (
	"crypto/ed25519"
	"path/filepath"
	"testing"

	"github.com/libr-forum/Libr/core/libr/dbnode"
	"github.com/libr-forum/Libr/core/libr/proto"
)

var _ dbnode.Store = (*Store)(nil)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, _ := ed25519.GenerateKey(nil)
	c := proto.SignMsg(priv, proto.Msg{Community: "general", Ts: 1800000000, Content: []byte("hi")})
	h := c.Hash()

	if got, err := s.Get(h); err != nil || got != nil {
		t.Fatalf("empty store: %v %v", got, err)
	}
	for i := 0; i < 2; i++ { // idempotent
		if err := s.Put(h, c); err != nil {
			t.Fatal(err)
		}
		if err := s.AddIndex("general", 900, h, c.Msg.ID()); err != nil {
			t.Fatal(err)
		}
	}
	s.Close()

	s, err = Open(path) // persists across restarts
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got, err := s.Get(h)
	if err != nil || got == nil || got.Hash() != h {
		t.Fatalf("get after reopen: %+v %v", got, err)
	}
	hs, err := s.Index("general", 900, 10)
	if err != nil || len(hs) != 1 || hs[0] != h {
		t.Fatalf("index: %v %v", hs, err)
	}
	if hs, _ := s.Index("general", 1800, 10); len(hs) != 0 {
		t.Fatalf("other bucket: %v", hs)
	}
}

// TestRemove: a tombstone deletes every cert and index entry of the message, survives
// a restart, and is listed for its bucket.
func TestRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, _ := ed25519.GenerateKey(nil)
	c := proto.SignMsg(priv, proto.Msg{Community: "general", Ts: 1800000000, Content: []byte("hi")})
	c2 := c
	c2.Epoch = 1 // a second certificate of the same message
	keep := proto.SignMsg(priv, proto.Msg{Community: "general", Ts: 1800000000, Content: []byte("keep")})
	for _, x := range []proto.MsgCert{c, c2, keep} {
		if err := s.Put(x.Hash(), x); err != nil {
			t.Fatal(err)
		}
		if err := s.AddIndex("general", 900, x.Hash(), x.Msg.ID()); err != nil {
			t.Fatal(err)
		}
	}
	del := proto.SignAuthorDelete(priv, c.Msg)
	tomb := proto.Tombstone{MsgID: c.Msg.ID(), Del: &del}
	if err := s.Remove(tomb, "general", 900); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if s, err = Open(path); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, x := range []proto.MsgCert{c, c2} {
		if got, _ := s.Get(x.Hash()); got != nil {
			t.Fatal("removed cert still stored")
		}
	}
	if got, _ := s.Get(keep.Hash()); got == nil {
		t.Fatal("unrelated cert deleted")
	}
	if hs, _ := s.Index("general", 900, 10); len(hs) != 1 || hs[0] != keep.Hash() {
		t.Fatalf("index after remove: %v", hs)
	}
	if ok, err := s.Removed(c.Msg.ID()); !ok || err != nil {
		t.Fatalf("Removed = %v %v", ok, err)
	}
	if ts, err := s.Tombstones("general", 900, 10); err != nil || len(ts) != 1 || ts[0].MsgID != c.Msg.ID() || ts[0].Kind() != "author" {
		t.Fatalf("tombstones: %+v %v", ts, err)
	}
}
