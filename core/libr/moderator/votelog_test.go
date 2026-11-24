package moderator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// TestVoteIsFinal: a policy that would approve on a second try (an LLM is not
// deterministic) never gets that second try. The first signed vote is returned to
// every repeat, also after a restart, so no author can retry its way to approval.
func TestVoteIsFinal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "votes.jsonl")
	calls := 0
	flaky := func(context.Context, proto.Msg) (bool, error) { calls++; return calls > 1, nil }
	m, _ := setup(t, flaky)
	m.VoteLog = path
	c := msg("a", "borderline", now)
	for i := 0; i < 3; i++ {
		v, err := m.Decide(context.Background(), c)
		if err != nil {
			t.Fatal(err)
		}
		if v.Approve {
			t.Fatalf("ask %d approved after a rejection", i)
		}
	}
	if calls != 1 {
		t.Fatalf("policy ran %d times, want 1", calls)
	}

	m2, _ := setup(t, ApproveAll) // restart with a policy that approves everything
	m2.VoteLog = path
	v, err := m2.Decide(context.Background(), c)
	if err != nil || v.Approve {
		t.Fatalf("after restart: vote %+v err %v, want the recorded rejection", v, err)
	}
	if _, err := m2.Decide(context.Background(), msg("a", "fresh", now)); err != nil {
		t.Fatal(err)
	}
}

// TestReportDecisionPerPost: reports with different reasons, from different
// reporters, map to one queue item and one decision, so neither the queue nor the
// policy can be flooded or re-rolled with new reason strings.
func TestReportDecisionPerPost(t *testing.T) {
	m, _ := setup(t, deferAll)
	q, _ := NewQueue("")
	m.Queue = q
	m.reporters = wire.NewLimiter(1000, 1000)
	cert := published("reported")
	for i := 0; i < 10; i++ {
		r := report(t, fmt.Sprint("reporter", i), cert, fmt.Sprint("reason ", i))
		if _, err := m.DecideReport(context.Background(), r); !errors.Is(err, ErrPending) {
			t.Fatalf("report %d: %v", i, err)
		}
	}
	items := q.Items()
	if len(items) != 1 {
		t.Fatalf("queued %d items, want 1", len(items))
	}
	if err := m.Resolve(items[0].Key, false); err != nil {
		t.Fatal(err)
	}
	r := report(t, "late", cert, "a brand new reason")
	v, err := m.DecideReport(context.Background(), r)
	if err != nil || v.Approve {
		t.Fatalf("after keep: vote %+v err %v", v, err)
	}
	rc := proto.RepCert{MsgID: cert.Msg.ID(), Community: "general", Reason: "a brand new reason"}
	if !proto.VerifyRepVote(*v, pub(key("mod0")), rc) {
		t.Fatal("vote not signed over the reason asked")
	}
}

// TestReportPolicyRunsOnce: a nondeterministic report policy is asked once per post.
func TestReportPolicyRunsOnce(t *testing.T) {
	m, _ := setup(t, ApproveAll)
	m.reporters = wire.NewLimiter(1000, 1000)
	calls := 0
	m.Report = func(context.Context, proto.Msg, string) (bool, error) { calls++; return calls > 1, nil }
	cert := published("borderline")
	for i := 0; i < 5; i++ {
		v, err := m.DecideReport(context.Background(), report(t, fmt.Sprint("r", i), cert, fmt.Sprint(i)))
		if err != nil || v.Approve {
			t.Fatalf("report %d: %+v %v", i, v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("report policy ran %d times", calls)
	}
}

// TestRejectedContentStaysRejected: re-signing rejected content with a new timestamp
// or a new author key does not buy a fresh policy decision.
func TestRejectedContentStaysRejected(t *testing.T) {
	calls := 0
	flaky := func(context.Context, proto.Msg) (bool, error) { calls++; return calls > 1, nil }
	m, _ := setup(t, flaky)
	for i, c := range []proto.MsgCert{
		msg("a", "same text", now),
		msg("a", "same text", now.Add(time.Second)),
		msg("someone-else", "same text", now),
	} {
		v, err := m.Decide(context.Background(), c)
		if err != nil || v.Approve {
			t.Fatalf("variant %d: %+v %v", i, v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("policy ran %d times", calls)
	}
	if v, err := m.Decide(context.Background(), msg("a", "other text", now)); err != nil || !v.Approve {
		t.Fatalf("other content: %+v %v", v, err)
	}
}

// TestNoEquivocationUnderConcurrency: concurrent identical requests while the policy
// sometimes defers never yield both a signed vote and a queued item, and a human
// decision can never contradict a signed vote.
func TestNoEquivocationUnderConcurrency(t *testing.T) {
	var mu sync.Mutex
	n := 0
	racy := func(context.Context, proto.Msg) (bool, error) {
		mu.Lock()
		n++
		k := n
		mu.Unlock()
		time.Sleep(time.Millisecond)
		if k%2 == 0 {
			return false, ErrDefer
		}
		return true, nil
	}
	for trial := 0; trial < 20; trial++ {
		m, _ := setup(t, racy)
		q, _ := NewQueue("")
		m.Queue = q
		m.authors = wire.NewLimiter(1000, 1000)
		c := msg("a", fmt.Sprint("post ", trial), now)
		var wg sync.WaitGroup
		approvals := 0
		var amu sync.Mutex
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if v, err := m.Decide(context.Background(), c); err == nil && v.Approve {
					amu.Lock()
					approvals++
					amu.Unlock()
				}
			}()
		}
		wg.Wait()
		items := q.Items()
		if approvals > 0 && len(items) > 0 {
			t.Fatalf("trial %d: signed approvals and queued the same post", trial)
		}
		for _, it := range items {
			if err := m.Resolve(it.Key, false); err != nil {
				t.Fatal(err)
			}
			if v, err := m.Decide(context.Background(), c); err != nil || v.Approve {
				t.Fatalf("trial %d: after human reject got %+v %v", trial, v, err)
			}
		}
	}
}

// TestConcurrentResignsShareOneDecision: many copies of the same content signed with
// different author keys, sent at once to a policy that approves after its first
// (rejecting) call, all get the first rejection.
func TestConcurrentResignsShareOneDecision(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	flaky := func(context.Context, proto.Msg) (bool, error) {
		mu.Lock()
		calls++
		k := calls
		mu.Unlock()
		time.Sleep(time.Millisecond)
		return k > 1, nil
	}
	m, _ := setup(t, flaky)
	m.authors = wire.NewLimiter(1000, 1000)
	var wg sync.WaitGroup
	var amu sync.Mutex
	approved := 0
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := m.Decide(context.Background(), msg(fmt.Sprint("author", i), "same borderline text", now))
			if err == nil && v.Approve {
				amu.Lock()
				approved++
				amu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if approved != 0 || calls != 1 {
		t.Fatalf("approved %d, policy calls %d", approved, calls)
	}
}
