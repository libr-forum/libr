package forum

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/proto"
)

// resolveAll makes mods[i] decide every queued item: approve[i].
func (w *world) resolveAll(t *testing.T, approve ...bool) {
	t.Helper()
	for i, ok := range approve {
		m := w.modObjs[i]
		for _, it := range m.Queue.Items() {
			if it.Vote == nil {
				if err := m.Resolve(it.Key, ok); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func asPending(t *testing.T, err error) *Pending {
	t.Helper()
	var qe *QuorumError
	if !errors.As(err, &qe) || !errors.Is(err, ErrPending) {
		t.Fatalf("err = %v, want a pending QuorumError", err)
	}
	return qe.Pending
}

// TestHumanReview: image posts and all reports wait for human moderators; the
// author's and reporter's clients resume and complete once a quorum has decided.
func TestHumanReview(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	images := func(_ context.Context, m proto.Msg) (bool, error) {
		if strings.Contains(string(m.Content), "<img") {
			return false, moderator.ErrDefer
		}
		return true, nil
	}
	humans := func(context.Context, proto.Msg, string) (bool, error) { return false, moderator.ErrDefer }
	w := newWorld(t, 1, 4, images, withQueue(), withReport(humans))
	ctx, cancel := ctx90()
	defer cancel()

	plain, err := w.submit("plain text is decided at once")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.submit(`<img src="data:image/png;base64,AAAA">`)
	p := asPending(t, err)
	if p.Waiting != 4 || len(p.Votes) != 0 {
		t.Fatalf("pending: waiting %d votes %d", p.Waiting, len(p.Votes))
	}
	if _, _, err := w.client.Resume(ctx, p); !errors.Is(err, ErrPending) {
		t.Fatalf("resume before review: %v", err)
	}
	w.resolveAll(t, true, false, true, true) // one mod rejects; 3 approvals suffice
	posted, _, err := w.client.Resume(ctx, p)
	if err != nil || posted == nil || len(posted.Cert.Votes) != 3 {
		t.Fatalf("resume after review: %+v %v", posted, err)
	}
	if got := w.feed(); len(got) != 2 {
		t.Fatalf("feed = %q", got)
	}

	// Report: queued; two moderators keep it, so removal is impossible.
	_, err = w.client.Report(ctx, posted.Cert, "dislike")
	rp := asPending(t, err)
	w.resolveAll(t, false, false)
	if _, _, err := w.client.Resume(ctx, rp); !errors.Is(err, ErrRejected) {
		t.Fatalf("report kept by 2 of 4: %v", err)
	}

	// A second report with another reason gets the same decisions: still kept.
	if _, err := w.client.Report(ctx, posted.Cert, "illegal"); !errors.Is(err, ErrRejected) {
		t.Fatalf("second report with a new reason: %v", err)
	}

	// Another post: three moderators remove it.
	_, err = w.client.Report(ctx, plain.Cert, "illegal")
	rp = asPending(t, err)
	w.resolveAll(t, true, true, true)
	_, removed, err := w.client.Resume(ctx, rp)
	if err != nil || removed == nil || removed.Tomb.Kind() != "mod" {
		t.Fatalf("report removed by 3 of 4: %+v %v", removed, err)
	}
	<-removed.Done
	if got := w.feed(); len(got) != 1 || got[0] == "plain text is decided at once" {
		t.Fatalf("feed after removal = %q", got)
	}
}
