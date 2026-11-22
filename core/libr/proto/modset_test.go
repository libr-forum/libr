package proto

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"testing"
	"testing/fstest"
)

func TestGenesisChecks(t *testing.T) {
	founder := key("general/founder")
	base := func() ModSet { return genesis(t, "general", 1).set.unverified() }
	sign := func(s ModSet, signers ...ed25519.PrivateKey) ModSet {
		s.Sigs = nil
		for _, k := range signers {
			SignModSet(k, &s)
		}
		return s
	}
	founders := []ed25519.PublicKey{pub(founder)}

	cases := []struct {
		name     string
		set      func() ModSet
		founders []ed25519.PublicKey
		ok       bool
	}{
		{"valid", func() ModSet { return sign(base(), founder) }, founders, true},
		{"too few mods for f", func() ModSet { s := base(); s.Mods = s.Mods[:3]; return sign(s, founder) }, founders, false},
		{"threshold not 2f+1", func() ModSet { s := base(); s.Threshold = 2; return sign(s, founder) }, founders, false},
		{"peer id does not match key", func() ModSet { s := base(); s.Mods[0].PeerID = s.Mods[1].PeerID; return sign(s, founder) }, founders, false},
		{"duplicate mod", func() ModSet { s := base(); s.Mods[1] = s.Mods[0]; return sign(s, founder) }, founders, false},
		{"unsigned", func() ModSet { return sign(base()) }, founders, false},
		{"signed by a stranger", func() ModSet { return sign(base(), key("stranger")) }, founders, false},
		{"extra signer", func() ModSet { return sign(base(), founder, key("stranger")) }, founders, false},
		{"missing a co-founder", func() ModSet { return sign(base(), founder) }, append(founders, pub(key("cofounder"))), false},
		{"both co-founders", func() ModSet { return sign(base(), founder, key("cofounder")) }, append(founders, pub(key("cofounder"))), true},
		{"tampered after signing", func() ModSet { s := sign(base(), founder); s.Mods = s.Mods[:len(s.Mods)-1]; s.F = 0; s.Threshold = 1; return s }, founders, false},
		{"non-zero prev hash", func() ModSet { s := base(); s.PrevHash = Sum([]byte("x")); return sign(s, founder) }, founders, false},
		{"genesis not epoch 0", func() ModSet { s := base(); s.Epoch = 3; return sign(s, founder) }, founders, false},
		{"no founders trusted", func() ModSet { return sign(base(), founder) }, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := VerifyModSetChain(tc.founders, []ModSet{tc.set()})
			if (err == nil) != tc.ok {
				t.Fatalf("ok=%v err=%v", tc.ok, err)
			}
		})
	}
}

func TestLaterEpochsFailClosed(t *testing.T) {
	g := genesis(t, "general", 1)
	next := g.set.unverified()
	next.Epoch, next.PrevHash = 1, g.set.Hash()
	_, err := VerifyModSetChain([]ed25519.PublicKey{pub(g.founder)}, []ModSet{g.set.unverified(), next})
	if !errors.Is(err, ErrHandoffNotImplemented) {
		t.Fatalf("err = %v, want ErrHandoffNotImplemented", err)
	}
}

func TestLoadRegistry(t *testing.T) {
	file := func(name string) []byte {
		g := genesis(t, name, 1)
		b, _ := json.Marshal(CommunityFile{Founders: []ed25519.PublicKey{pub(g.founder)}, Chain: []ModSet{g.set.unverified()}})
		return b
	}
	r, err := LoadRegistry(fstest.MapFS{
		"general.json": {Data: file("general")},
		"crypto.json":  {Data: file("crypto")},
		"notes.txt":    {Data: []byte("ignored")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Communities(); len(got) != 2 || got[0] != "crypto" || got[1] != "general" {
		t.Fatalf("communities = %v", got)
	}
	if s, ok := r.Get("general", 0); !ok || !s.IsMember(pub(key("general/mod0"))) {
		t.Fatal("general epoch 0 not loaded with members")
	}
	if _, ok := r.Get("general", 1); ok {
		t.Fatal("unknown epoch found")
	}

	if _, err := LoadRegistry(fstest.MapFS{"a.json": {Data: file("general")}, "b.json": {Data: file("general")}}); err == nil {
		t.Fatal("community defined twice accepted")
	}
	bad := file("general")
	bad[len(bad)/2] ^= 1
	if _, err := LoadRegistry(fstest.MapFS{"general.json": {Data: bad}}); err == nil {
		t.Fatal("corrupted file accepted")
	}
}

// unverified copies s without its member index, as it would arrive from disk.
func (s *ModSet) unverified() ModSet {
	out := *s
	out.members = nil
	out.Mods = append([]ModEntry(nil), s.Mods...)
	out.Sigs = append([]Sig(nil), s.Sigs...)
	return out
}
