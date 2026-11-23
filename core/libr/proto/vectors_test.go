package proto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// TestGoldenVectors pins the exact bytes every signature covers. A change here breaks
// signatures across the network; regenerate deliberately with UPDATE_GOLDEN=1.
func TestGoldenVectors(t *testing.T) {
	c := genesis(t, "general", 1)
	author := key("author")
	mc := c.cert(author, "hello libr", c.mods[0], c.mods[1], c.mods[2])
	rc := RepCert{MsgID: mc.Msg.ID(), Community: "general", Reason: "spam"}
	r := SignReceipt(key("db1"), mc, 1_800_000_100)
	h := func(b []byte) string { return hex.EncodeToString(b) }
	id, ch, sh := mc.Msg.ID(), mc.Hash(), c.set.Hash()

	got := map[string]string{
		"author_pub":         h(pub(author)),
		"msg_bytes":          h(mc.Msg.Bytes()),
		"msg_id":             id.String(),
		"author_sig":         h(mc.AuthorSig),
		"modvote_bytes":      h(ModVoteBytes("general", 0, id, true)),
		"modvote_sig_mod0":   h(mc.Votes[0].Sig),
		"msgcert_hash":       ch.String(),
		"repvote_bytes":      h(RepVoteBytes("general", 0, rc.MsgID, rc.Reason, true)),
		"authordelete_bytes": h(AuthorDeleteBytes(id)),
		"receipt_bytes":      h(r.Bytes()),
		"receipt_sig":        h(r.Sig),
		"modset_hash":        sh.String(),
		"mod0_peer_id":       c.set.Mods[0].PeerID,
	}
	b, _ := json.MarshalIndent(got, "", "  ")
	b = append(b, '\n')
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile("testdata/vectors.json", b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, want) {
		t.Fatalf("signing bytes changed:\n%s\nwant:\n%s", b, want)
	}
}

func TestEncodeLengthPrefixed(t *testing.T) {
	// The v1 scheme signed content+ts+status, so ("ab","c") and ("a","bc") collided.
	if bytes.Equal(Encode("d", []byte("ab"), []byte("c")), Encode("d", []byte("a"), []byte("bc"))) {
		t.Fatal("field boundaries are ambiguous")
	}
	if bytes.Equal(Encode("libr/v2/x", nil), Encode("libr/v2/x")) {
		t.Fatal("empty field and no field encode the same")
	}
}

// FuzzEncodeInjective: different field lists never encode to the same bytes.
func FuzzEncodeInjective(f *testing.F) {
	f.Add("d", []byte("ab"), []byte("c"), "d", []byte("a"), []byte("bc"))
	f.Add("libr/v2/msg", []byte{}, []byte{1}, "libr/v2/ms", []byte("g"), []byte{1})
	f.Fuzz(func(t *testing.T, d1 string, a1, b1 []byte, d2 string, a2, b2 []byte) {
		same := d1 == d2 && bytes.Equal(a1, a2) && bytes.Equal(b1, b2)
		if !same && bytes.Equal(Encode(d1, a1, b1), Encode(d2, a2, b2)) {
			t.Fatalf("collision: %q %x %x vs %q %x %x", d1, a1, b1, d2, a2, b2)
		}
	})
}

// FuzzVerifyMsgCert: flipping any byte of any signed field of a valid cert makes it
// fail verification.
func FuzzVerifyMsgCert(f *testing.F) {
	c := genesis(f, "general", 1)
	valid := c.cert(key("author"), "hello libr", c.mods[0], c.mods[1], c.mods[2])
	if err := VerifyMsgCert(c.reg, valid, testNow); err != nil {
		f.Fatal(err)
	}
	f.Add(uint8(0), uint16(0), uint8(1))
	f.Add(uint8(5), uint16(31), uint8(0x80))
	f.Fuzz(func(t *testing.T, field uint8, idx uint16, xor uint8) {
		if xor == 0 {
			return
		}
		mc := valid
		mc.Msg.Content = append([]byte(nil), valid.Msg.Content...)
		mc.Msg.Author = append([]byte(nil), valid.Msg.Author...)
		mc.AuthorSig = append([]byte(nil), valid.AuthorSig...)
		mc.Votes = append([]Vote(nil), valid.Votes...)
		v := &mc.Votes[int(field)%len(mc.Votes)]
		v.Mod, v.Sig = append([]byte(nil), v.Mod...), append([]byte(nil), v.Sig...)
		targets := [][]byte{mc.Msg.Content, mc.Msg.Author, mc.AuthorSig, v.Mod, v.Sig}
		buf := targets[int(field)%len(targets)]
		buf[int(idx)%len(buf)] ^= xor
		if err := VerifyMsgCert(c.reg, mc, testNow); err == nil {
			t.Fatalf("mutated cert accepted (field %d idx %d xor %x)", field, idx, xor)
		}
	})
}

// FuzzDecodeMsgCert: arbitrary wire bytes never panic decoding or verification.
func FuzzDecodeMsgCert(f *testing.F) {
	c := genesis(f, "general", 1)
	b, _ := json.Marshal(c.cert(key("author"), "hello", c.mods[0], c.mods[1], c.mods[2]))
	f.Add(b)
	f.Add([]byte(`{"msg":{"community":"general"},"votes":[{"mod":"","approve":true}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var mc MsgCert
		if json.Unmarshal(data, &mc) == nil {
			_ = VerifyMsgCert(c.reg, mc, testNow)
			_ = mc.Hash()
		}
	})
}
