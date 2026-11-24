// Package moderator serves /libr/2/vote: a community moderator checks an author-signed
// message, runs its content policy and answers with a signed approve or reject vote.
package moderator

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// Policy decides whether a message may be published. An error means the policy could
// not decide (e.g. the LLM is unreachable), and the message is rejected. A moderator
// asks its policy at most once per message: the vote it signs is final.
type Policy func(ctx context.Context, m proto.Msg) (bool, error)

// ApproveAll approves everything (testbeds).
func ApproveAll(context.Context, proto.Msg) (bool, error) { return true, nil }

// Keywords rejects messages containing any of words, case-insensitively.
func Keywords(words ...string) Policy {
	return func(_ context.Context, m proto.Msg) (bool, error) {
		text := strings.ToLower(string(m.Content))
		for _, w := range words {
			if w != "" && strings.Contains(text, strings.ToLower(w)) {
				return false, nil
			}
		}
		return true, nil
	}
}

// ReportPolicy decides whether a reported message should be removed. An error means
// it could not decide, and the vote is against removal.
type ReportPolicy func(ctx context.Context, m proto.Msg, reason string) (bool, error)

// RemoveIfRejected removes exactly what p would not have approved.
func RemoveIfRejected(p Policy) ReportPolicy {
	return func(ctx context.Context, m proto.Msg, _ string) (bool, error) {
		ok, err := p(ctx, m)
		return !ok && err == nil, err
	}
}

// Request budgets. Per peer bounds any one connection; per author bounds one key
// spamming through many peers, and is charged only after the author signature checks
// out, so nobody can exhaust someone else's budget.
const (
	peerPerSecond   = 5
	peerBurst       = 20
	authorPerSecond = 0.2
	authorBurst     = 5
	decideTimeout   = 20 * time.Second // under wire.Timeout, so the reply still arrives
	// maxDeciding bounds concurrent policy calls (an LLM is the expensive part); peer
	// IDs and author keys are free to mint, so per-key limits alone do not bound it.
	maxDeciding = 16
)

// How long decisions are kept. A post vote outlives the timestamp window by
// clockMargin; a rejected content hash is kept for contentFor; a report decision is
// kept for as long as the moderator runs (reports have no timestamp window).
var (
	clockMargin = time.Hour
	contentFor  = 30 * 24 * time.Hour
	forever     = time.Unix(1<<40, 0)
)

// ErrBusy is returned when maxDeciding decisions are already running.
var ErrBusy = errors.New("busy")

// Mod is one moderator.
type Mod struct {
	Key    ed25519.PrivateKey
	Sets   proto.ModSets
	Policy Policy
	Report ReportPolicy // nil means RemoveIfRejected(Policy)
	// Queue receives items a policy defers (ErrDefer) for a human to decide with
	// Resolve. Nil: deferrals are rejections.
	Queue *Queue
	Now   func() time.Time // nil means time.Now
	// VoteLog is the file of every vote this moderator signed; a repeated request
	// gets the same vote. Empty: kept in memory only (a restart forgets it).
	VoteLog string

	locksMu   sync.Mutex
	locks     map[string]*keyLock
	votesOnce sync.Once
	votes     *votes
	votesErr  error

	authors   *wire.Limiter
	reporters *wire.Limiter
	slots     chan struct{}
}

// decide runs fn in one of maxDeciding slots, refusing at once when all are taken.
func (m *Mod) decide(fn func()) error {
	if m.slots == nil { // Decide used without Register (tests)
		fn()
		return nil
	}
	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
		fn()
		return nil
	default:
		return ErrBusy
	}
}

// Register installs /libr/2/vote on h.
func (m *Mod) Register(h host.Host) {
	m.authors = wire.NewLimiter(authorPerSecond, authorBurst)
	m.reporters = wire.NewLimiter(authorPerSecond, authorBurst)
	m.slots = make(chan struct{}, maxDeciding)
	peers := wire.NewLimiter(peerPerSecond, peerBurst)
	reject := func(err error) wire.VoteResp { return wire.VoteResp{Err: err.Error()} }
	wire.Handle(h, wire.ProtoVote, peers, m.vote, reject)
	wire.Handle(h, wire.ProtoReport, peers, func(_ peer.ID, r wire.ReportReq) wire.VoteResp {
		return voteResp(m.DecideReport(context.Background(), r))
	}, reject)
}

func (m *Mod) vote(_ peer.ID, c wire.VoteReq) wire.VoteResp {
	return voteResp(m.Decide(context.Background(), c))
}

func voteResp(v *proto.Vote, err error) wire.VoteResp {
	switch {
	case errors.Is(err, ErrPending):
		return wire.VoteResp{Pending: true}
	case err != nil:
		return wire.VoteResp{Err: err.Error()}
	}
	return wire.VoteResp{Vote: v}
}

// lock serializes everything about one decision key (a post or a report on a post),
// so two concurrent requests cannot both run the policy, or one sign while the other
// queues the item: a moderator's decision on a key is made once.
func (m *Mod) lock(key string) func() {
	m.locksMu.Lock()
	if m.locks == nil {
		m.locks = map[string]*keyLock{}
	}
	l := m.locks[key]
	if l == nil {
		l = &keyLock{}
		m.locks[key] = l
	}
	l.refs++
	m.locksMu.Unlock()
	l.mu.Lock()
	return func() {
		l.mu.Unlock()
		m.locksMu.Lock()
		if l.refs--; l.refs == 0 {
			delete(m.locks, key)
		}
		m.locksMu.Unlock()
	}
}

type keyLock struct {
	mu   sync.Mutex
	refs int
}

// voteLog opens the vote log once.
func (m *Mod) voteLog() (*votes, error) {
	m.votesOnce.Do(func() { m.votes, m.votesErr = openVotes(m.VoteLog, m.now()) })
	return m.votes, m.votesErr
}

// decided returns this moderator's decision on key: from the vote log, or from a
// human decision in the queue. pending is true while a human has not decided yet.
func (m *Mod) decided(key string) (approve, ok, pending bool, err error) {
	log, err := m.voteLog()
	if err != nil {
		return false, false, false, err
	}
	if v, found := log.get(key, m.now()); found {
		return v.Approve, true, false, nil
	}
	if m.Queue != nil {
		if it, found := m.Queue.get(key); found {
			if it.Vote == nil {
				return false, false, true, nil
			}
			return it.Vote.Approve, true, false, nil
		}
	}
	return false, false, false, nil
}

// record makes a decision final: it is written (and synced) to the vote log before any
// vote carrying it leaves this moderator.
func (m *Mod) record(key string, until time.Time, v proto.Vote) error {
	log, err := m.voteLog()
	if err != nil {
		return err
	}
	_, err = log.put(key, until, v, m.now())
	return err
}

// Resolve records a human decision on a queued item. Every later request for the
// item receives a vote with this decision.
func (m *Mod) Resolve(key string, approve bool) error {
	unlock := m.lock(key)
	defer unlock()
	it, ok := m.Queue.get(key)
	if !ok {
		return fmt.Errorf("no queued item %q", key)
	}
	if _, done, _, err := m.decided(key); err != nil {
		return err
	} else if done {
		return errors.New("already decided")
	}
	var v proto.Vote
	until := m.now().Add(keepFor)
	switch it.Kind {
	case "post":
		v = proto.SignModVote(m.Key, it.Cert.Msg, it.Epoch, approve)
		if !approve {
			if err := m.record(ContentKey(it.Cert.Msg, it.Epoch), m.now().Add(contentFor), v); err != nil {
				return err
			}
		}
	case "report":
		v = proto.SignRepVote(m.Key, repCert(it.Cert.Msg, it.Epoch, it.Reason), approve)
		until = forever
	default:
		return fmt.Errorf("unknown item kind %q", it.Kind)
	}
	if err := m.record(key, until, v); err != nil {
		return err
	}
	return m.Queue.decide(key, v)
}

func repCert(msg proto.Msg, epoch uint64, reason string) proto.RepCert {
	return proto.RepCert{MsgID: msg.ID(), Community: msg.Community, Epoch: epoch, Reason: reason}
}

func (m *Mod) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

// member checks this moderator belongs to community's set for epoch.
func (m *Mod) member(community string, epoch uint64) error {
	set, ok := m.Sets.Get(community, epoch)
	if !ok {
		return errors.New("unknown community or epoch")
	}
	if !set.IsMember(m.Key.Public().(ed25519.PublicKey)) {
		return errors.New("not a moderator of this community")
	}
	return nil
}

// Decide checks c and returns this moderator's signed vote on it.
func (m *Mod) Decide(ctx context.Context, c proto.MsgCert) (*proto.Vote, error) {
	now := m.now()
	msg := c.Msg
	if err := m.member(msg.Community, c.Epoch); err != nil {
		return nil, err
	}
	if err := proto.VerifyAuthor(c); err != nil {
		return nil, err
	}
	// A decided or queued post is answered before the timestamp check: the author
	// re-asks later, and gets the same decision.
	key := PostKey(msg, c.Epoch)
	unlock := m.lock(key)
	defer unlock()
	sign := func(approve bool) *proto.Vote { v := proto.SignModVote(m.Key, msg, c.Epoch, approve); return &v }
	if approve, ok, pending, err := m.decided(key); err != nil {
		return nil, err
	} else if pending {
		return nil, ErrPending
	} else if ok {
		return sign(approve), nil
	}
	// Both directions: a backdated message would land in an old feed bucket.
	if d := now.Sub(time.Unix(msg.Ts, 0)); d > proto.MaxClockSkew || d < -proto.MaxClockSkew {
		return nil, errors.New("timestamp too far from now")
	}
	// Content this moderator rejected before stays rejected, whoever signs it and
	// whenever: re-signing a post must not buy a fresh policy decision.
	ckey := ContentKey(msg, c.Epoch)
	// Held across the policy call: concurrent re-signed copies of the same content
	// (new timestamps or author keys) wait, then see a rejection recorded by the first.
	unlockContent := m.lock(ckey)
	defer unlockContent()
	if approve, ok, _, err := m.decided(ckey); err != nil {
		return nil, err
	} else if ok && !approve {
		v := sign(false)
		return v, m.record(key, postUntil(msg), *v)
	}
	if m.authors != nil && !m.authors.Allow(string(msg.Author)) {
		return nil, wire.ErrRateLimited
	}
	ctx, cancel := context.WithTimeout(ctx, decideTimeout)
	defer cancel()
	var ok bool
	var err error
	if busy := m.decide(func() { ok, err = m.Policy(ctx, msg) }); busy != nil {
		return nil, busy
	}
	if errors.Is(err, ErrDefer) && m.Queue != nil {
		if qerr := m.Queue.add(Item{Key: key, Kind: "post", Cert: proto.MsgCert{Msg: msg, AuthorSig: c.AuthorSig}, Epoch: c.Epoch}); qerr != nil {
			return nil, qerr
		}
		return nil, ErrPending
	}
	v := sign(ok && err == nil)
	if !v.Approve {
		if err := m.record(ckey, m.now().Add(contentFor), *v); err != nil {
			return nil, err
		}
	}
	return v, m.record(key, postUntil(msg), *v)
}

// postUntil is how long a post vote is kept: past the end of the window in which the
// timestamp check admits the post, with a margin for clock adjustments.
func postUntil(msg proto.Msg) time.Time {
	return time.Unix(msg.Ts, 0).Add(proto.MaxClockSkew + clockMargin)
}

// DecideReport checks a report and returns this moderator's signed vote on removing
// the message. Only published (verifying) certificates can be reported, and every
// report is signed by its reporter, who is rate-limited.
func (m *Mod) DecideReport(ctx context.Context, r wire.ReportReq) (*proto.Vote, error) {
	msg := r.Cert.Msg
	if err := m.member(msg.Community, r.Epoch); err != nil {
		return nil, err
	}
	if len(r.Reason) > proto.MaxReason {
		return nil, errors.New("reason too long")
	}
	if err := proto.VerifyMsgCert(m.Sets, r.Cert, m.now()); err != nil {
		return nil, fmt.Errorf("reported certificate does not verify: %w", err)
	}
	id := msg.ID()
	if len(r.Reporter) != ed25519.PublicKeySize || len(r.Sig) != ed25519.SignatureSize ||
		!ed25519.Verify(r.Reporter, proto.ReportBytes(msg.Community, r.Epoch, id, r.Reason), r.Sig) {
		return nil, errors.New("bad reporter signature")
	}
	// One decision per reported post, whatever the reason: a reporter cannot buy a fresh
	// policy decision with a new reason string. The vote is signed over the reason asked.
	key := ReportKey(msg, r.Epoch)
	unlock := m.lock(key)
	defer unlock()
	sign := func(remove bool) *proto.Vote {
		v := proto.SignRepVote(m.Key, repCert(msg, r.Epoch, r.Reason), remove)
		return &v
	}
	if remove, ok, pending, err := m.decided(key); err != nil {
		return nil, err
	} else if pending {
		return nil, ErrPending
	} else if ok {
		return sign(remove), nil
	}
	if m.reporters != nil && !m.reporters.Allow(string(r.Reporter)) {
		return nil, wire.ErrRateLimited
	}
	policy := m.Report
	if policy == nil {
		policy = RemoveIfRejected(m.Policy)
	}
	ctx, cancel := context.WithTimeout(ctx, decideTimeout)
	defer cancel()
	var remove bool
	var err error
	if busy := m.decide(func() { remove, err = policy(ctx, msg, r.Reason) }); busy != nil {
		return nil, busy
	}
	if errors.Is(err, ErrDefer) && m.Queue != nil {
		if qerr := m.Queue.add(Item{Key: key, Kind: "report", Cert: r.Cert, Epoch: r.Epoch, Reason: r.Reason}); qerr != nil {
			return nil, qerr
		}
		return nil, ErrPending
	}
	v := sign(remove && err == nil)
	return v, m.record(key, forever, *v)
}
