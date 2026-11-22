package moderator

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

func key(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("moderator-test/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

func pub(k ed25519.PrivateKey) ed25519.PublicKey { return k.Public().(ed25519.PublicKey) }

var now = time.Unix(1_800_000_000, 0)

func setup(t *testing.T, policy Policy) (*Mod, *proto.Registry) {
	t.Helper()
	s := proto.ModSet{Community: "general", F: 1, Threshold: 3}
	for i := 0; i < 4; i++ {
		k := key(fmt.Sprint("mod", i))
		id, _ := proto.PeerIDFor(pub(k))
		s.Mods = append(s.Mods, proto.ModEntry{PubKey: pub(k), PeerID: id})
	}
	proto.SignModSet(key("founder"), &s)
	reg := &proto.Registry{}
	if err := reg.Add(proto.CommunityFile{Founders: []ed25519.PublicKey{pub(key("founder"))}, Chain: []proto.ModSet{s}}); err != nil {
		t.Fatal(err)
	}
	m := &Mod{Key: key("mod0"), Sets: reg, Policy: policy, Now: func() time.Time { return now }}
	m.authors = wire.NewLimiter(authorPerSecond, authorBurst)
	return m, reg
}

func msg(author string, content string, ts time.Time) proto.MsgCert {
	return proto.SignMsg(key(author), proto.Msg{Community: "general", Ts: ts.Unix(), Content: []byte(content)})
}

func TestDecide(t *testing.T) {
	failing := func(context.Context, proto.Msg) (bool, error) { return true, errors.New("llm down") }
	for _, tc := range []struct {
		name    string
		policy  Policy
		cert    func() proto.MsgCert
		approve bool
		err     bool
	}{
		{"approve", ApproveAll, func() proto.MsgCert { return msg("a", "hello", now) }, true, false},
		{"keyword reject", Keywords("Spam"), func() proto.MsgCert { return msg("a", "SPAM here", now) }, false, false},
		{"policy error rejects", failing, func() proto.MsgCert { return msg("a", "hello", now) }, false, false},
		{"unknown community", ApproveAll, func() proto.MsgCert {
			return proto.SignMsg(key("a"), proto.Msg{Community: "other", Ts: now.Unix(), Content: []byte("x")})
		}, false, true},
		{"unknown epoch", ApproveAll, func() proto.MsgCert { c := msg("a", "x", now); c.Epoch = 1; return c }, false, true},
		{"future timestamp", ApproveAll, func() proto.MsgCert { return msg("a", "x", now.Add(10*time.Minute)) }, false, true},
		{"backdated", ApproveAll, func() proto.MsgCert { return msg("a", "x", now.Add(-10*time.Minute)) }, false, true},
		{"bad author signature", ApproveAll, func() proto.MsgCert { c := msg("a", "x", now); c.Msg.Content = []byte("y"); return c }, false, true},
		{"empty content", ApproveAll, func() proto.MsgCert { return msg("a", "", now) }, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := setup(t, tc.policy)
			c := tc.cert()
			v, err := m.Decide(context.Background(), c)
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got vote %+v", v)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if v.Approve != tc.approve || !proto.VerifyVote(*v, pub(key("mod0")), c.Msg, c.Epoch) {
				t.Fatalf("vote approve=%v valid=%v, want approve=%v", v.Approve, proto.VerifyVote(*v, pub(key("mod0")), c.Msg, c.Epoch), tc.approve)
			}
		})
	}
}

func TestNonMemberRefuses(t *testing.T) {
	m, _ := setup(t, ApproveAll)
	m.Key = key("outsider")
	if _, err := m.Decide(context.Background(), msg("a", "x", now)); err == nil {
		t.Fatal("a non-member signed a vote")
	}
}

// TestAuthorRateLimit: one author is limited however many peers they use; another
// author is unaffected.
func TestAuthorRateLimit(t *testing.T) {
	m, _ := setup(t, ApproveAll)
	limited := false
	for i := 0; i < authorBurst+1; i++ {
		if _, err := m.Decide(context.Background(), msg("spammer", fmt.Sprint(i), now)); errors.Is(err, wire.ErrRateLimited) {
			limited = true
		}
	}
	if !limited {
		t.Fatalf("author not limited after %d requests", authorBurst+1)
	}
	if _, err := m.Decide(context.Background(), msg("someone-else", "hi", now)); err != nil {
		t.Fatalf("other author: %v", err)
	}
}
