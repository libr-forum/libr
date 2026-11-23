//go:build devnet

package main

// Drives the desktop adapter against a running dev network (testbed/devnet.sh) with a
// throwaway identity: run with HOME=<empty dir> and the devnet's LIBR_* environment.
//
//	HOME=$(mktemp -d) LIBR_RELAYS=... LIBR_BOOTSTRAP=... LIBR_COMMUNITIES=... \
//	  go test -tags 'devnet webkit2_41' -run TestDevnet -v .

import (
	"strings"
	"testing"
	"time"
)

func TestDevnet(t *testing.T) {
	a := NewApp()
	if err := a.Connect(); err != nil {
		t.Fatal(err)
	}
	defer a.stop()
	st := a.GetNetworkStatus()
	t.Logf("status %+v", st)
	if a.GetRelayStatus() != "online" || st.Reservations == 0 {
		t.Fatalf("not online behind the relay: %+v", st)
	}
	cs, err := a.ListCommunities()
	if err != nil || len(cs) != 1 || cs[0].Name != "general" || cs[0].IsMod {
		t.Fatalf("communities %+v %v", cs, err)
	}

	text := "hello from the desktop adapter " + time.Now().Format(time.RFC3339Nano)
	start := time.Now()
	out, err := a.SendInput("general", text)
	if err != nil || out.Status != "posted" {
		t.Fatalf("send: %+v %v", out, err)
	}
	t.Logf("posted in %v", time.Since(start).Round(time.Millisecond))
	img := `<BODY><img src="data:image/png;base64,iVBORw0KGgo=" /></BODY>`
	if out, err := a.SendInput("general", img); err != nil || out.Status != "posted" {
		t.Fatalf("image (devnet mods approve all): %+v %v", out, err)
	}

	posts, err := a.FetchAll("general")
	if err != nil {
		t.Fatal(err)
	}
	var mine string
	for _, p := range posts {
		if p.Content == text && p.Mine {
			mine = p.Hash
		}
	}
	if mine == "" {
		t.Fatalf("own post missing from feed of %d", len(posts))
	}
	// devnet mods remove only posts containing "remove-me": this report is rejected.
	if r, err := a.Report(mine, "testing reports"); err != nil || r.Status != "rejected" {
		t.Fatalf("report: %+v %v", r, err)
	}
	if err := a.Delete(mine); err != nil {
		t.Fatal(err)
	}
	posts, _ = a.FetchAll("general")
	for _, p := range posts {
		if strings.Contains(p.Content, text) {
			t.Fatal("deleted post still in feed")
		}
	}
	if p := a.GetPending(); len(p) != 0 {
		t.Fatalf("unexpected pending %+v", p)
	}
}
