package forum

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// Kinds of Pending.
const (
	KindPost   = "post"
	KindReport = "report"
)

// ErrPending: moderators queued the item for human review; Resume later.
var ErrPending = errors.New("waiting for moderators")

// Pending is a post or report still collecting moderator votes. It is plain data so
// apps can keep it across restarts and Resume it.
type Pending struct {
	Kind   string        `json:"kind"`
	Cert   proto.MsgCert `json:"cert"` // the post (unvoted for KindPost)
	Epoch  uint64        `json:"epoch"`
	Reason string        `json:"reason,omitempty"`
	Votes  []proto.Vote  `json:"votes,omitempty"` // decisions so far, approvals and rejections
	// Waiting is how many moderators answered "pending" in the last round.
	Waiting int `json:"waiting"`
}

// Tally counts the decisions so far.
func (p *Pending) Tally() (approved, rejected int) {
	for _, v := range p.Votes {
		if v.Approve {
			approved++
		} else {
			rejected++
		}
	}
	return
}

// QuorumError reports that no quorum was reached yet. It unwraps to ErrPending
// (moderators are reviewing) or ErrNoQuorum (too few answered); Pending can be
// resumed either way.
type QuorumError struct {
	Pending *Pending
	Err     error
}

func (e *QuorumError) Error() string { return e.Err.Error() }
func (e *QuorumError) Unwrap() error { return e.Err }

// Resume asks the moderators that have not decided p yet and, once a quorum is in,
// completes the post (Posted) or the removal (Removed).
func (c *Client) Resume(ctx context.Context, p *Pending) (*Posted, *Removed, error) {
	return c.advance(ctx, p)
}

func (c *Client) advance(ctx context.Context, p *Pending) (*Posted, *Removed, error) {
	m := p.Cert.Msg
	set, ok := c.Sets.Get(m.Community, p.Epoch)
	if !ok {
		return nil, nil, fmt.Errorf("unknown community %q epoch %d", m.Community, p.Epoch)
	}
	var ask func(context.Context, proto.ModEntry) (*proto.Vote, error)
	ballot := proto.RepCert{MsgID: m.ID(), Community: m.Community, Epoch: p.Epoch, Reason: p.Reason}
	switch p.Kind {
	case KindPost:
		unvoted := p.Cert // late answers still read it after quorum returns
		ask = func(ctx context.Context, mod proto.ModEntry) (*proto.Vote, error) {
			return c.askVote(ctx, mod, unvoted)
		}
	case KindReport:
		req := wire.ReportReq{Cert: p.Cert, Epoch: p.Epoch, Reason: p.Reason, Reporter: c.Key.Public().(ed25519.PublicKey),
			Sig: ed25519.Sign(c.Key, proto.ReportBytes(m.Community, p.Epoch, ballot.MsgID, p.Reason))}
		ask = func(ctx context.Context, mod proto.ModEntry) (*proto.Vote, error) {
			return c.askReport(ctx, mod, req, ballot)
		}
	default:
		return nil, nil, fmt.Errorf("unknown pending kind %q", p.Kind)
	}

	start := time.Now()
	votes, waiting, err := c.quorum(ctx, set, p.Votes, ask)
	c.span("quorum", start, err)
	p.Votes, p.Waiting = votes, waiting
	if errors.Is(err, ErrPending) || errors.Is(err, ErrNoQuorum) {
		return nil, nil, &QuorumError{Pending: p, Err: err}
	}
	if err != nil {
		return nil, nil, err
	}
	var approvals []proto.Vote
	for _, v := range votes {
		if v.Approve {
			approvals = append(approvals, v)
		}
	}
	if p.Kind == KindPost {
		cert := p.Cert
		cert.Votes = approvals
		posted, err := c.finishPost(ctx, cert)
		return posted, nil, err
	}
	ballot.Votes = approvals
	if err := proto.VerifyRepCert(c.Sets, ballot); err != nil {
		return nil, nil, fmt.Errorf("assembled removal certificate does not verify: %w", err)
	}
	r, err := c.remove(ctx, m, p.Cert.Hash(), proto.Tombstone{MsgID: ballot.MsgID, Rep: &ballot})
	return nil, r, err
}

// quorum asks, in parallel, every moderator without a decision in prior. It returns
// all decisions (prior plus new) as soon as Threshold approvals are in, ErrRejected as
// soon as rejections make that impossible, and otherwise ErrPending (some moderators
// are reviewing) or ErrNoQuorum once everyone has answered or failed.
func (c *Client) quorum(ctx context.Context, set *proto.ModSet, prior []proto.Vote, ask func(context.Context, proto.ModEntry) (*proto.Vote, error)) ([]proto.Vote, int, error) {
	n, need := len(set.Mods), int(set.Threshold)
	votes := append([]proto.Vote(nil), prior...)
	approved, rejected := 0, 0
	decided := func(k ed25519.PublicKey) bool {
		for _, v := range votes {
			if bytes.Equal(v.Mod, k) {
				return true
			}
		}
		return false
	}
	for _, v := range votes {
		if v.Approve {
			approved++
		} else {
			rejected++
		}
	}
	if approved >= need {
		return votes, 0, nil
	}
	if n-rejected < need {
		return votes, 0, ErrRejected
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type answer struct {
		v   *proto.Vote
		err error
	}
	var todo []proto.ModEntry
	for _, m := range set.Mods {
		if !decided(m.PubKey) {
			todo = append(todo, m)
		}
	}
	answers := make(chan answer, len(todo))
	for _, m := range todo {
		go func(m proto.ModEntry) {
			v, err := ask(ctx, m)
			answers <- answer{v, err}
		}(m)
	}
	waiting, failed := 0, 0
	for range todo {
		a := <-answers
		switch {
		case errors.Is(a.err, ErrPending):
			waiting++
		case a.err != nil:
			failed++
		case a.v.Approve:
			approved++
			votes = append(votes, *a.v)
		default:
			rejected++
			votes = append(votes, *a.v)
		}
		if approved >= need {
			return votes, 0, nil
		}
		if n-rejected < need {
			return votes, 0, ErrRejected
		}
	}
	if waiting > 0 {
		return votes, waiting, ErrPending
	}
	return votes, 0, fmt.Errorf("%w: %d approved, %d rejected, %d unreachable, need %d", ErrNoQuorum, approved, rejected, failed, need)
}
