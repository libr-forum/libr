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
		if err := s.AddIndex("general", 900, h); err != nil {
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
