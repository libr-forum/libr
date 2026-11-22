// Package moderator serves /libr/2/vote: a community moderator checks an author-signed
// message, runs its content policy and answers with a signed approve or reject vote.
package moderator

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// Policy decides whether a message may be published. An error means the policy could
// not decide (e.g. the LLM is unreachable), and the message is rejected.
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

// ErrBusy is returned when maxDeciding decisions are already running.
var ErrBusy = errors.New("busy")

// Mod is one moderator.
type Mod struct {
	Key    ed25519.PrivateKey
	Sets   proto.ModSets
	Policy Policy
	Report ReportPolicy     // nil means RemoveIfRejected(Policy)
	Now    func() time.Time // nil means time.Now

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
		v, err := m.DecideReport(context.Background(), r)
		if err != nil {
			return wire.VoteResp{Err: err.Error()}
		}
		return wire.VoteResp{Vote: v}
	}, reject)
}

func (m *Mod) vote(_ peer.ID, c wire.VoteReq) wire.VoteResp {
	v, err := m.Decide(context.Background(), c)
	if err != nil {
		return wire.VoteResp{Err: err.Error()}
	}
	return wire.VoteResp{Vote: v}
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
	// Both directions: a backdated message would land in an old feed bucket.
	if d := now.Sub(time.Unix(msg.Ts, 0)); d > proto.MaxClockSkew || d < -proto.MaxClockSkew {
		return nil, errors.New("timestamp too far from now")
	}
	if err := proto.VerifyAuthor(c); err != nil {
		return nil, err
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
	v := proto.SignModVote(m.Key, msg, c.Epoch, ok && err == nil)
	return &v, nil
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
	v := proto.SignRepVote(m.Key, proto.RepCert{MsgID: id, Community: msg.Community, Epoch: r.Epoch, Reason: r.Reason}, remove && err == nil)
	return &v, nil
}
