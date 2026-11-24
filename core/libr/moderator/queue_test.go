package moderator

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
)

func deferAll(context.Context, proto.Msg) (bool, error) { return false, ErrDefer }

// TestDeferResolve: a deferred post is queued and answered "pending" until a human
// decides; then every re-ask gets the signed vote, even long after the message's
// timestamp window, and the decision survives a restart.
func TestDeferResolve(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	q, err := NewQueue(path)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := setup(t, deferAll)
	m.Queue = q
	c := msg("a", "an image post", now)

	for i := 0; i < 2; i++ {
		if _, err := m.Decide(context.Background(), c); !errors.Is(err, ErrPending) {
			t.Fatalf("ask %d: err = %v, want ErrPending", i, err)
		}
	}
	items := q.Items()
	if len(items) != 1 || items[0].Kind != "post" || items[0].Vote != nil {
		t.Fatalf("queue = %+v", items)
	}
	if err := m.Resolve(items[0].Key, true); err != nil {
		t.Fatal(err)
	}
	if err := m.Resolve(items[0].Key, false); err == nil {
		t.Fatal("decided twice")
	}

	later := now.Add(3 * time.Hour)
	m.Now = func() time.Time { return later }
	q2, err := NewQueue(path) // restart
	if err != nil {
		t.Fatal(err)
	}
	q2.now = func() time.Time { return later }
	m.Queue = q2
	v, err := m.Decide(context.Background(), c)
	if err != nil || !v.Approve || !proto.VerifyVote(*v, pub(key("mod0")), c.Msg, 0) {
		t.Fatalf("after resolve: %+v %v", v, err)
	}
	// A different post is still subject to the timestamp window.
	if _, err := m.Decide(context.Background(), msg("a", "new", now)); err == nil || errors.Is(err, ErrPending) {
		t.Fatalf("stale new post: %v", err)
	}
}

func TestDeferReport(t *testing.T) {
	m, _ := setup(t, ApproveAll)
	m.Report = func(context.Context, proto.Msg, string) (bool, error) { return false, ErrDefer }
	m.Queue, _ = NewQueue("")
	req := report(t, "r", published("hello"), "spam")
	if _, err := m.DecideReport(context.Background(), req); !errors.Is(err, ErrPending) {
		t.Fatalf("err = %v", err)
	}
	it := m.Queue.Items()[0]
	if it.Kind != "report" || it.Reason != "spam" {
		t.Fatalf("item %+v", it)
	}
	if err := m.Resolve(it.Key, true); err != nil {
		t.Fatal(err)
	}
	v, err := m.DecideReport(context.Background(), req)
	r := proto.RepCert{MsgID: req.Cert.Msg.ID(), Community: "general", Reason: "spam"}
	if err != nil || !v.Approve || !proto.VerifyRepVote(*v, pub(key("mod0")), r) {
		t.Fatalf("after resolve: %+v %v", v, err)
	}
	// Same message, different reason: the same decision, signed over the new reason.
	v, err = m.DecideReport(context.Background(), report(t, "r2", published("hello"), "other"))
	r.Reason = "other"
	if err != nil || !v.Approve || !proto.VerifyRepVote(*v, pub(key("mod0")), r) {
		t.Fatalf("second reason: %+v %v", v, err)
	}
}

func TestNoQueueDeferRejects(t *testing.T) {
	m, _ := setup(t, deferAll)
	v, err := m.Decide(context.Background(), msg("a", "x", now))
	if err != nil || v.Approve {
		t.Fatalf("without a queue a deferral must reject: %+v %v", v, err)
	}
}
