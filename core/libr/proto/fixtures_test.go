package proto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

// key derives a deterministic Ed25519 key from a label, so vectors are reproducible.
func key(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("libr-proto-test/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

type community struct {
	reg      *Registry
	set      *ModSet
	founder  ed25519.PrivateKey
	mods     []ed25519.PrivateKey
	outsider ed25519.PrivateKey
}

// genesis builds a verified registry for a community of 3f+1 moderators.
func genesis(t testing.TB, name string, f int) *community {
	t.Helper()
	c := &community{founder: key(name + "/founder"), outsider: key(name + "/outsider")}
	s := ModSet{Community: name, Epoch: 0, F: uint64(f), Threshold: uint64(2*f + 1)}
	for i := 0; i < 3*f+1; i++ {
		k := key(fmt.Sprintf("%s/mod%d", name, i))
		c.mods = append(c.mods, k)
		id, err := PeerIDFor(pub(k))
		if err != nil {
			t.Fatal(err)
		}
		s.Mods = append(s.Mods, ModEntry{PubKey: pub(k), PeerID: id})
	}
	SignModSet(c.founder, &s)
	c.reg = &Registry{}
	if err := c.reg.Add(CommunityFile{Founders: []ed25519.PublicKey{pub(c.founder)}, Chain: []ModSet{s}}); err != nil {
		t.Fatal(err)
	}
	c.set, _ = c.reg.Get(name, 0)
	return c
}

var testNow = time.Unix(1_800_000_000, 0)

// cert returns a message by author approved by the given moderators.
func (c *community) cert(author ed25519.PrivateKey, content string, approvers ...ed25519.PrivateKey) MsgCert {
	mc := SignMsg(author, Msg{Community: c.set.Community, Ts: testNow.Unix(), Content: []byte(content)})
	for _, m := range approvers {
		mc.Votes = append(mc.Votes, SignModVote(m, mc.Msg, mc.Epoch, true))
	}
	return mc
}
