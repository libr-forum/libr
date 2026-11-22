package proto

import (
	"crypto/ed25519"
	"math/rand"
	"testing"
	"time"
)

func TestVerifyMsgCert(t *testing.T) {
	c := genesis(t, "general", 1) // 4 mods, threshold 3
	other := genesis(t, "other", 1)
	author := key("author")
	m := c.mods

	cases := []struct {
		name string
		mut  func(*MsgCert)
		ok   bool
	}{
		{"threshold approvals", func(*MsgCert) {}, true},
		{"all mods", func(mc *MsgCert) { mc.Votes = append(mc.Votes, SignModVote(m[3], mc.Msg, 0, true)) }, true},
		{"below threshold", func(mc *MsgCert) { mc.Votes = mc.Votes[:2] }, false},
		{"duplicate vote fills quorum", func(mc *MsgCert) { mc.Votes[2] = mc.Votes[0] }, false},
		{"self-approval by outsider", func(mc *MsgCert) { mc.Votes[2] = SignModVote(c.outsider, mc.Msg, 0, true) }, false},
		{"rejection in cert", func(mc *MsgCert) { mc.Votes = append(mc.Votes, SignModVote(m[3], mc.Msg, 0, false)) }, false},
		{"approve flag flipped on a reject sig", func(mc *MsgCert) { v := SignModVote(m[2], mc.Msg, 0, false); v.Approve = true; mc.Votes[2] = v }, false},
		{"vote from another epoch", func(mc *MsgCert) { mc.Votes[2] = SignModVote(m[2], mc.Msg, 1, true) }, false},
		{"cert claims unknown epoch", func(mc *MsgCert) { mc.Epoch = 1 }, false},
		{"vote for another community", func(mc *MsgCert) {
			alt := mc.Msg
			alt.Community = "other"
			mc.Votes[2] = SignModVote(m[2], alt, 0, true)
		}, false},
		{"moved to another community", func(mc *MsgCert) { mc.Msg.Community = "other" }, false},
		{"unknown community", func(mc *MsgCert) { mc.Msg.Community = "nope" }, false},
		{"content edited after signing", func(mc *MsgCert) { mc.Msg.Content = []byte("edited") }, false},
		{"author swapped", func(mc *MsgCert) { mc.Msg.Author = pub(key("someone")) }, false},
		{"bad author sig length", func(mc *MsgCert) { mc.AuthorSig = mc.AuthorSig[:10] }, false},
		{"future timestamp", func(mc *MsgCert) { *mc = c.reSign(author, mc, testNow.Add(10*time.Minute).Unix()) }, false},
		{"slight clock skew ok", func(mc *MsgCert) { *mc = c.reSign(author, mc, testNow.Add(time.Minute).Unix()) }, true},
		{"empty content", func(mc *MsgCert) { mc.Msg.Content = nil }, false},
		{"more votes than mods", func(mc *MsgCert) {
			for i := 0; i < 3; i++ {
				mc.Votes = append(mc.Votes, mc.Votes[0])
			}
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mc := c.cert(author, "hello", m[0], m[1], m[2])
			tc.mut(&mc)
			reg := multi{c.reg, other.reg}
			if err := VerifyMsgCert(reg, mc, testNow); (err == nil) != tc.ok {
				t.Fatalf("ok=%v err=%v", tc.ok, err)
			}
		})
	}
}

// reSign rebuilds mc with a new timestamp, re-signed by author and the same voters.
func (c *community) reSign(author ed25519.PrivateKey, mc *MsgCert, ts int64) MsgCert {
	out := SignMsg(author, Msg{Community: mc.Msg.Community, Ts: ts, Content: mc.Msg.Content})
	for _, v := range mc.Votes {
		for _, m := range c.mods {
			if string(pub(m)) == string(v.Mod) {
				out.Votes = append(out.Votes, SignModVote(m, out.Msg, out.Epoch, true))
			}
		}
	}
	return out
}

// multi looks up a set in several registries.
type multi []*Registry

func (ms multi) Get(c string, e uint64) (*ModSet, bool) {
	for _, r := range ms {
		if s, ok := r.Get(c, e); ok {
			return s, ok
		}
	}
	return nil, false
}

// TestQuorumProperty: for f in 1..3, any 2f+1 distinct members pass, any 2f fail,
// and 2f+1 votes drawn from only 2f keys fail.
func TestQuorumProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	author := key("author")
	for f := 1; f <= 3; f++ {
		c := genesis(t, "prop", f)
		for trial := 0; trial < 50; trial++ {
			perm := rng.Perm(len(c.mods))
			pick := func(n int) []ed25519.PrivateKey {
				var out []ed25519.PrivateKey
				for _, i := range perm[:n] {
					out = append(out, c.mods[i])
				}
				return out
			}
			if err := VerifyMsgCert(c.reg, c.cert(author, "x", pick(2*f+1)...), testNow); err != nil {
				t.Fatalf("f=%d: 2f+1 distinct members rejected: %v", f, err)
			}
			if err := VerifyMsgCert(c.reg, c.cert(author, "x", pick(2*f)...), testNow); err == nil {
				t.Fatalf("f=%d: 2f members accepted", f)
			}
			padded := c.cert(author, "x", pick(2*f)...)
			padded.Votes = append(padded.Votes, padded.Votes[rng.Intn(len(padded.Votes))])
			if err := VerifyMsgCert(c.reg, padded, testNow); err == nil {
				t.Fatalf("f=%d: 2f+1 votes from 2f keys accepted", f)
			}
		}
	}
}

// TestNoConflictingCerts: with at most f equivocating moderators, approve and reject
// quorums for the same message cannot both form, since 2(2f+1) - (3f+1) = f+1 > f.
func TestNoConflictingCerts(t *testing.T) {
	for f := 1; f <= 3; f++ {
		n := 3*f + 1
		q := 2*f + 1
		if overlap := 2*q - n; overlap <= f {
			t.Fatalf("f=%d: quorum overlap %d does not exceed f", f, overlap)
		}
	}
}

func TestCertHashIgnoresVoteOrder(t *testing.T) {
	c := genesis(t, "general", 1)
	a := c.cert(key("author"), "hi", c.mods[0], c.mods[1], c.mods[2])
	b := a
	b.Votes = []Vote{a.Votes[2], a.Votes[0], a.Votes[1]}
	if a.Hash() != b.Hash() {
		t.Fatal("hash depends on vote order")
	}
	b.Votes = b.Votes[:2]
	if a.Hash() == b.Hash() {
		t.Fatal("hash ignores the vote set")
	}
}

func TestVerifyRepCert(t *testing.T) {
	c := genesis(t, "general", 1)
	id := c.cert(key("author"), "bad", c.mods[0]).Msg.ID()
	rc := RepCert{MsgID: id, Community: "general", Reason: "spam"}
	for _, m := range c.mods[:3] {
		rc.Votes = append(rc.Votes, SignRepVote(m, rc, true))
	}
	if err := VerifyRepCert(c.reg, rc); err != nil {
		t.Fatal(err)
	}
	short := rc
	short.Votes = rc.Votes[:2]
	if VerifyRepCert(c.reg, short) == nil {
		t.Fatal("2f removal votes accepted")
	}
	reworded := rc
	reworded.Reason = "harassment"
	if VerifyRepCert(c.reg, reworded) == nil {
		t.Fatal("reason changed after voting accepted")
	}
	retarget := rc
	retarget.MsgID = Sum([]byte("other"))
	if VerifyRepCert(c.reg, retarget) == nil {
		t.Fatal("votes reused for another message")
	}
	// A publish vote must not count as a removal vote.
	m := c.cert(key("author"), "bad", c.mods[0]).Msg
	cross := rc
	cross.Votes = []Vote{SignModVote(c.mods[0], m, 0, true), rc.Votes[1], rc.Votes[2]}
	if VerifyRepCert(c.reg, cross) == nil {
		t.Fatal("publish vote accepted as removal vote")
	}
}

func TestAuthorDelete(t *testing.T) {
	c := genesis(t, "general", 1)
	author := key("author")
	m := c.cert(author, "mine").Msg
	if err := VerifyAuthorDelete(SignAuthorDelete(author, m), m); err != nil {
		t.Fatal(err)
	}
	if VerifyAuthorDelete(SignAuthorDelete(key("mallory"), m), m) == nil {
		t.Fatal("someone else deleted the post")
	}
	other := c.cert(author, "other post").Msg
	if VerifyAuthorDelete(SignAuthorDelete(author, other), m) == nil {
		t.Fatal("delete for another message accepted")
	}
}

func TestStoreReceipt(t *testing.T) {
	c := genesis(t, "general", 1)
	r := SignReceipt(key("db1"), c.cert(key("author"), "x", c.mods[:3]...), 123)
	if !r.Verify() {
		t.Fatal("valid receipt rejected")
	}
	forged := r
	forged.Node = pub(key("db2"))
	if forged.Verify() {
		t.Fatal("receipt accepted under another node's key")
	}
	moved := r
	moved.Community = "other"
	if moved.Verify() {
		t.Fatal("receipt community changed")
	}
}
