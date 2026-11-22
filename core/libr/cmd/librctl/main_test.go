package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// TestKeygenGenesis: keys and a genesis file made by the CLI load into a registry,
// and keygen is stable for an existing key file.
func TestKeygenGenesis(t *testing.T) {
	dir := t.TempDir()
	keygen := func(name string) []string {
		var out bytes.Buffer
		if err := run([]string{"keygen", "-out", filepath.Join(dir, name+".key")}, &out); err != nil {
			t.Fatal(err)
		}
		return strings.Fields(out.String())
	}
	var mods []string
	for _, m := range []string{"m1", "m2", "m3", "m4"} {
		mods = append(mods, keygen(m)[1])
	}
	if again := keygen("m1"); again[1] != mods[0] {
		t.Fatal("keygen changed an existing key")
	}
	keygen("founder")
	var out bytes.Buffer
	args := []string{"genesis", "-community", "general", "-f", "1", "-founders", filepath.Join(dir, "founder.key"), "-mods", strings.Join(mods, ",")}
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}
	comm := filepath.Join(dir, "communities")
	if err := os.Mkdir(comm, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(comm, "general.json"), out.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	reg, err := proto.LoadRegistry(os.DirFS(comm))
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := reg.Head("general"); !ok || s.Threshold != 3 || len(s.Mods) != 4 {
		t.Fatalf("registry head = %+v", s)
	}
	if err := run(append(args[:len(args)-1], mods[0]), &out); err == nil {
		t.Fatal("genesis accepted 1 mod for f=1")
	}
}

// TestSeparateFounderSigning: founders sign one by one on their own machines; the
// file verifies only once every founder has signed, and outsiders cannot sign.
func TestSeparateFounderSigning(t *testing.T) {
	dir := t.TempDir()
	keygen := func(name string) string {
		var out bytes.Buffer
		if err := run([]string{"keygen", "-out", filepath.Join(dir, name+".key")}, &out); err != nil {
			t.Fatal(err)
		}
		return strings.Fields(out.String())[1]
	}
	var mods []string
	for _, m := range []string{"m1", "m2", "m3", "m4"} {
		mods = append(mods, keygen(m))
	}
	a, b := keygen("alice"), keygen("bob")
	keygen("mallory")
	file := filepath.Join(dir, "general.json")
	do := func(args ...string) error { var out bytes.Buffer; return run(args, &out) }
	if err := do("genesis", "-community", "general", "-f", "1", "-mods", strings.Join(mods, ","), "-founder-keys", a+","+b, "-out", file); err != nil {
		t.Fatal(err)
	}
	if do("verify", "-in", file) == nil {
		t.Fatal("unsigned file verified")
	}
	if do("sign", "-key", filepath.Join(dir, "mallory.key"), "-in", file) == nil {
		t.Fatal("a non-founder signed")
	}
	if err := do("sign", "-key", filepath.Join(dir, "alice.key"), "-in", file); err != nil {
		t.Fatal(err)
	}
	if do("sign", "-key", filepath.Join(dir, "alice.key"), "-in", file) == nil {
		t.Fatal("signed twice")
	}
	if do("verify", "-in", file) == nil {
		t.Fatal("verified with 1 of 2 founders")
	}
	if err := do("sign", "-key", filepath.Join(dir, "bob.key"), "-in", file); err != nil {
		t.Fatal(err)
	}
	if err := do("verify", "-in", file); err != nil {
		t.Fatalf("fully signed file: %v", err)
	}
	reg, err := proto.LoadRegistry(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.Head("general"); !ok {
		t.Fatal("not loadable by nodes")
	}
}
