package peer

import (
	"crypto/ed25519"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	lpeer "github.com/libp2p/go-libp2p/core/peer"
)

func TestStableIdentity(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	lpub, err := crypto.UnmarshalEd25519PublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := lpeer.IDFromPublicKey(lpub)

	for i := 0; i < 2; i++ {
		id, err := StableIdentity(priv)
		if err != nil {
			t.Fatal(err)
		}
		h, err := libp2p.New(id, libp2p.NoListenAddrs)
		if err != nil {
			t.Fatal(err)
		}
		if h.ID() != want {
			t.Fatalf("run %d: peer ID %s, want %s (derived from app key)", i, h.ID(), want)
		}
		_ = h.Close()
	}
	if _, err := StableIdentity(priv[:10]); err == nil {
		t.Fatal("short key accepted")
	}
}

func TestServePostReqDispatch(t *testing.T) {
	t.Cleanup(func() {
		postMu.Lock()
		postRoutes, postFallback = map[string]PostHandler{}, nil
		postMu.Unlock()
	})
	params := func(route string) []byte { b, _ := json.Marshal(map[string]string{"route": route}); return b }

	if got := ServePostReq("a", params("auto"), []byte(`{}`)); got != nil {
		t.Fatalf("unregistered route answered %q", got)
	}
	HandlePost("auto", func(from string, body []byte) []byte { return []byte("auto:" + from) })
	HandleUnknownPost(func(string, []byte) []byte { return []byte("fallback") })

	if got := string(ServePostReq("a", params("auto"), []byte(`{}`))); got != "auto:a" {
		t.Fatalf("auto -> %q", got)
	}
	if got := string(ServePostReq("a", params("other"), []byte(`{}`))); got != "fallback" {
		t.Fatalf("other -> %q", got)
	}
	if got := ServePostReq("a", []byte("not json"), nil); got != nil {
		t.Fatalf("bad params answered %q", got)
	}
}
