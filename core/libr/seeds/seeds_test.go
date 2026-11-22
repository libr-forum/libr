package seeds

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// fakeServer implements librserver's register/peers contract (the Node app is tested
// against the same message format in services/librserver/test).
func fakeServer(t *testing.T) *httptest.Server {
	type reg struct {
		PublicKey, PeerID, Nonce, Signature string
		Addrs                               []string
	}
	var entries []entry
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/challenge", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"nonce": "n1"})
	})
	mux.HandleFunc("/nodes/register", func(w http.ResponseWriter, r *http.Request) {
		var b reg
		_ = json.NewDecoder(r.Body).Decode(&b)
		pk, _ := base64.StdEncoding.DecodeString(b.PublicKey)
		sig, _ := base64.StdEncoding.DecodeString(b.Signature)
		id, _ := proto.PeerIDFor(pk)
		if b.Nonce != "n1" || id != b.PeerID || !ed25519.Verify(pk, RegisterMessage(b.Nonce, b.PeerID, b.Addrs), sig) {
			http.Error(w, "bad", http.StatusUnauthorized)
			return
		}
		entries = append(entries, entry{b.PeerID, b.Addrs})
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "ttlSeconds": 600})
	})
	mux.HandleFunc("/peers", func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(entries) })
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func TestRegisterFetchResolve(t *testing.T) {
	s := fakeServer(t)
	ctx := context.Background()
	_, key, _ := ed25519.GenerateKey(nil)
	addrs := []ma.Multiaddr{ma.StringCast("/ip4/127.0.0.1/tcp/1"), ma.StringCast("/ip4/10.0.0.1/tcp/4001")}
	ttl, err := Register(ctx, s.URL, key, addrs)
	if err != nil || ttl.Seconds() != 600 {
		t.Fatalf("register: %v %v", ttl, err)
	}
	got, err := Fetch(ctx, s.URL, 8)
	if err != nil || len(got) != 1 || len(got[0].Addrs) != 1 || got[0].Addrs[0].String() != "/ip4/10.0.0.1/tcp/4001" {
		t.Fatalf("fetch (loopback must be dropped): %v %v", got, err)
	}

	// Config first, then cache, then the server; same peer's addresses merge.
	cfg := Parse([]string{"/ip4/10.0.0.9/tcp/1/p2p/" + got[0].ID.String()})
	cache := filepath.Join(t.TempDir(), "peers.json")
	_, other, _ := ed25519.GenerateKey(nil)
	oid, _ := proto.PeerIDFor(other.Public().(ed25519.PublicKey))
	_ = os.WriteFile(cache, []byte(`["/ip4/10.0.0.2/tcp/1/p2p/`+oid+`", "garbage"]`), 0o600)
	all := Resolve(ctx, cfg, cache, []string{s.URL}, 8)
	if len(all) != 2 || all[0].ID != got[0].ID || len(all[0].Addrs) != 2 || all[1].ID.String() != oid {
		t.Fatalf("resolve: %v", all)
	}
	// Enough peers already: the server is not asked.
	if all := Resolve(ctx, cfg, "", []string{"http://127.0.0.1:1"}, 1); len(all) != 1 {
		t.Fatalf("resolve without server: %v", all)
	}
}

func TestRegisterNeedsDialableAddrs(t *testing.T) {
	s := fakeServer(t)
	_, key, _ := ed25519.GenerateKey(nil)
	if _, err := Register(context.Background(), s.URL, key, []ma.Multiaddr{ma.StringCast("/ip4/127.0.0.1/tcp/1")}); err == nil {
		t.Fatal("registered loopback only")
	}
	if got, _ := Fetch(context.Background(), s.URL, 8); len(got) != 0 {
		t.Fatalf("server has %v", got)
	}
}
