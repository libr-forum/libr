package proto

import (
	"testing"
)

func TestVerifyTombstone(t *testing.T) {
	c := genesis(t, "general", 1)
	other := genesis(t, "other", 1)
	author := key("author")
	m := c.cert(author, "post", c.mods[0], c.mods[1], c.mods[2]).Msg

	rep := func(set *community, community string, msgID Hash, voters ...int) *RepCert {
		r := RepCert{MsgID: msgID, Community: community, Reason: "spam"}
		for _, i := range voters {
			r.Votes = append(r.Votes, SignRepVote(set.mods[i], r, true))
		}
		return &r
	}
	del := SignAuthorDelete(author, m)
	strangerDel := SignAuthorDelete(key("stranger"), m)
	otherMsg := c.cert(author, "another post").Msg
	both := Tombstone{MsgID: m.ID(), Rep: rep(c, "general", m.ID(), 0, 1, 2), Del: &del}
	sets := multi{c.reg, other.reg}

	for _, tc := range []struct {
		name string
		t    Tombstone
		ok   bool
	}{
		{"mod quorum 2f+1", Tombstone{MsgID: m.ID(), Rep: rep(c, "general", m.ID(), 0, 1, 2)}, true},
		{"only 2f mod votes", Tombstone{MsgID: m.ID(), Rep: rep(c, "general", m.ID(), 0, 1)}, false},
		{"another community's quorum", Tombstone{MsgID: m.ID(), Rep: rep(other, "other", m.ID(), 0, 1, 2)}, false},
		{"rep cert for another message", Tombstone{MsgID: m.ID(), Rep: rep(c, "general", otherMsg.ID(), 0, 1, 2)}, false},
		{"author delete", Tombstone{MsgID: m.ID(), Del: &del}, true},
		{"delete signed by someone else", Tombstone{MsgID: m.ID(), Del: &strangerDel}, false},
		{"tombstone id mismatch", Tombstone{MsgID: otherMsg.ID(), Del: &del}, false},
		{"two proofs", both, false},
		{"no proof", Tombstone{MsgID: m.ID()}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyTombstone(sets, tc.t, m)
			if (err == nil) != tc.ok {
				t.Fatalf("err = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

func TestRemoveAck(t *testing.T) {
	a := SignRemoveAck(key("db"), Sum([]byte("m")), 5)
	if !a.Verify() {
		t.Fatal("valid ack rejected")
	}
	a.Ts++
	if a.Verify() {
		t.Fatal("tampered ack accepted")
	}
}
