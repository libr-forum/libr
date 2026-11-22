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

func report(t *testing.T, reporter string, c proto.MsgCert, reason string) wire.ReportReq {
	t.Helper()
	k := key(reporter)
	return wire.ReportReq{Cert: c, Reason: reason, Reporter: pub(k),
		Sig: ed25519.Sign(k, proto.ReportBytes(c.Msg.Community, 0, c.Msg.ID(), reason))}
}

func published(content string) proto.MsgCert {
	c := msg("a", content, now)
	for i := 0; i < 3; i++ {
		c.Votes = append(c.Votes, proto.SignModVote(key(fmt.Sprint("mod", i)), c.Msg, 0, true))
	}
	return c
}

func TestDecideReport(t *testing.T) {
	failing := func(context.Context, proto.Msg) (bool, error) { return true, errors.New("llm down") }
	unpublished := msg("a", "never approved", now)
	forged := report(t, "r", published("spam"), "spam")
	forged.Reason = "something else"
	for _, tc := range []struct {
		name   string
		policy Policy
		req    wire.ReportReq
		remove bool
		err    bool
	}{
		{"remove what policy rejects", Keywords("spam"), report(t, "r", published("buy spam"), "spam"), true, false},
		{"keep what policy approves", Keywords("spam"), report(t, "r", published("hello"), "spam"), false, false},
		{"policy error keeps the post", failing, report(t, "r", published("hello"), "spam"), false, false},
		{"unpublished cert", ApproveAll, report(t, "r", unpublished, "spam"), false, true},
		{"reporter signature over other reason", ApproveAll, forged, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := setup(t, tc.policy)
			m.reporters = wire.NewLimiter(authorPerSecond, authorBurst)
			v, err := m.DecideReport(context.Background(), tc.req)
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got %+v", v)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			r := proto.RepCert{MsgID: tc.req.Cert.Msg.ID(), Community: "general", Reason: tc.req.Reason}
			if v.Approve != tc.remove || !proto.VerifyRepVote(*v, pub(key("mod0")), r) {
				t.Fatalf("vote %+v, want remove=%v", v, tc.remove)
			}
		})
	}
}
