package storage

import (
	"crypto/ed25519"
	"encoding/json"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/libr-forum/Libr/core/libr/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/models"
)

type key struct {
	pub  string
	priv ed25519.PrivateKey
}

func newKey(t *testing.T) key {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pub, _, err := cryptoutils.SignMessage(priv, "")
	if err != nil {
		t.Fatal(err)
	}
	return key{pub, priv}
}

func (k key) sign(t *testing.T, msg string) string {
	t.Helper()
	_, s, err := cryptoutils.SignMessage(k.priv, msg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func modSet(keys ...key) []*models.Mod {
	var out []*models.Mod
	for _, k := range keys {
		out = append(out, &models.Mod{PublicKey: k.pub})
	}
	return out
}

func vote(t *testing.T, k key, msg models.Msg, status string) models.ModCert {
	return models.ModCert{PublicKey: k.pub, Status: status, Sign: k.sign(t, msg.Content+strconv.FormatInt(msg.Ts, 10)+status)}
}

// signedCert builds a MsgCert whose outer author signature is valid.
func signedCert(t *testing.T, author key, msg models.Msg, votes ...models.ModCert) *models.MsgCert {
	sort.SliceStable(votes, func(i, j int) bool { return votes[i].PublicKey < votes[j].PublicKey })
	b, err := json.Marshal(models.DataToSign{Content: msg.Content, Ts: msg.Ts, ModCerts: votes})
	if err != nil {
		t.Fatal(err)
	}
	return &models.MsgCert{PublicKey: author.pub, Msg: msg, ModCerts: votes, Sign: author.sign(t, string(b))}
}

func TestValidateModCert(t *testing.T) {
	m1, m2, m3 := newKey(t), newKey(t), newKey(t)
	outsider := newKey(t)
	msg := models.Msg{Content: "hello", Ts: 1700000000}
	mods := modSet(m1, m2, m3)

	forged := vote(t, m2, msg, "1")
	forged.Sign = vote(t, m2, models.Msg{Content: "other", Ts: msg.Ts}, "1").Sign

	cases := []struct {
		name  string
		votes []models.ModCert
		mods  []*models.Mod
		ok    bool
	}{
		{"majority approves", []models.ModCert{vote(t, m1, msg, "1"), vote(t, m2, msg, "1")}, mods, true},
		{"self-approval by non-member", []models.ModCert{vote(t, outsider, msg, "1")}, mods, false},
		{"empty mod set fails closed", []models.ModCert{vote(t, m1, msg, "1")}, nil, false},
		{"self-approval with empty mod set", []models.ModCert{vote(t, outsider, msg, "1")}, nil, false},
		{"duplicate votes from one mod", []models.ModCert{vote(t, m1, msg, "1"), vote(t, m1, msg, "1"), vote(t, m1, msg, "1")}, mods, false},
		{"forged signature", []models.ModCert{vote(t, m1, msg, "1"), forged}, mods, false},
		{"no votes", nil, mods, false},
		{"one approval of three mods passes (v1 30% rule)", []models.ModCert{vote(t, m1, msg, "1")}, mods, true},
		{"rejections outweigh", []models.ModCert{vote(t, m1, msg, "1"), vote(t, m2, msg, "0"), vote(t, m3, msg, "0")}, mods, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateModCert(&models.MsgCert{Msg: msg, ModCerts: c.votes}, c.mods)
			if (err == nil) != c.ok {
				t.Fatalf("ok=%v, err=%v", c.ok, err)
			}
		})
	}
}

func TestValidateRepModCerts(t *testing.T) {
	m1, m2, m3 := newKey(t), newKey(t), newKey(t)
	author, other := newKey(t), newKey(t)
	msg := models.Msg{Content: "bad post", Ts: 1700000000}
	cert := signedCert(t, author, msg, vote(t, m1, msg, "1"), vote(t, m2, msg, "1"))
	repVote := func(k key, status string) models.ModCert {
		return models.ModCert{PublicKey: k.pub, Status: status, Sign: k.sign(t, cert.Sign+status)}
	}

	cases := []struct {
		name string
		rc   models.ReportCert
		mods []*models.Mod
		ok   bool
	}{
		{"report majority", models.ReportCert{Msgcert: *cert, Mode: "report", RepModCerts: []models.ModCert{repVote(m1, "1"), repVote(m2, "1")}}, modSet(m1, m2, m3), true},
		{"report duplicate votes", models.ReportCert{Msgcert: *cert, Mode: "report", RepModCerts: []models.ModCert{repVote(m1, "1"), repVote(m1, "1")}}, modSet(m1, m2, m3), false},
		{"report by non-member", models.ReportCert{Msgcert: *cert, Mode: "report", RepModCerts: []models.ModCert{repVote(other, "1")}}, modSet(m1, m2, m3), false},
		{"report with empty mod set", models.ReportCert{Msgcert: *cert, Mode: "report", RepModCerts: []models.ModCert{repVote(m1, "1")}}, nil, false},
		{"author delete", models.ReportCert{Msgcert: *cert, Mode: "delete", RepModCerts: []models.ModCert{{PublicKey: author.pub, Sign: author.sign(t, cert.Sign)}}}, nil, true},
		{"delete by someone else", models.ReportCert{Msgcert: *cert, Mode: "delete", RepModCerts: []models.ModCert{{PublicKey: other.pub, Sign: other.sign(t, cert.Sign)}}}, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateRepCert(&c.rc, c.mods)
			if (err == nil) != c.ok {
				t.Fatalf("ok=%v, err=%v", c.ok, err)
			}
		})
	}
}

func TestStoreAndDelete(t *testing.T) {
	db, err := config.OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	config.DB = db
	t.Cleanup(func() { db.Close() })

	m1, m2, m3 := newKey(t), newKey(t), newKey(t)
	author, outsider := newKey(t), newKey(t)
	mods := modSet(m1, m2, m3)
	msg := models.Msg{Content: "hi", Ts: 1700000000}

	if _, err := StoreMsgCert(signedCert(t, author, msg, vote(t, outsider, msg, "1")), mods); err == nil {
		t.Fatal("self-approved cert was stored")
	}
	if n := count(t, "SELECT COUNT(*) FROM msgcert"); n != 0 {
		t.Fatalf("rows after refused store = %d", n)
	}

	good := signedCert(t, author, msg, vote(t, m1, msg, "1"), vote(t, m2, msg, "1"))
	if _, err := StoreMsgCert(good, mods); err != nil {
		t.Fatal(err)
	}
	// A second post by the same author in the same second must survive the delete.
	msg2 := models.Msg{Content: "second", Ts: msg.Ts}
	if _, err := StoreMsgCert(signedCert(t, author, msg2, vote(t, m1, msg2, "1"), vote(t, m2, msg2, "1")), mods); err != nil {
		t.Fatal(err)
	}

	if err := DeleteMsgCert(&models.ReportCert{Msgcert: *good, Mode: "delete"}); err != nil {
		t.Fatal(err)
	}
	if n := count(t, "SELECT COUNT(*) FROM msgcert WHERE deleted = 1"); n != 1 {
		t.Fatalf("deleted rows = %d, want 1", n)
	}
}

func count(t *testing.T, q string) int {
	t.Helper()
	var n int
	if err := config.DB.QueryRow(q).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
