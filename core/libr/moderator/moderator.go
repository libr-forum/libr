// Package moderator serves /libr/2/vote: a community moderator checks an author-signed
// message, runs its content policy and answers with a signed approve or reject vote.
package moderator

import (
	"context"
	"crypto/ed25519"
	"errors"
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

// Request budgets. Per peer bounds any one connection; per author bounds one key
// spamming through many peers, and is charged only after the author signature checks
// out, so nobody can exhaust someone else's budget.
const (
	peerPerSecond   = 5
	peerBurst       = 20
	authorPerSecond = 0.2
	authorBurst     = 5
	decideTimeout   = 20 * time.Second // under wire.Timeout, so the reply still arrives
)

// Mod is one moderator.
type Mod struct {
	Key    ed25519.PrivateKey
	Sets   proto.ModSets
	Policy Policy
	Now    func() time.Time // nil means time.Now

	authors *wire.Limiter
}

// Register installs /libr/2/vote on h.
func (m *Mod) Register(h host.Host) {
	m.authors = wire.NewLimiter(authorPerSecond, authorBurst)
	wire.Handle(h, wire.ProtoVote, wire.NewLimiter(peerPerSecond, peerBurst), m.vote,
		func(err error) wire.VoteResp { return wire.VoteResp{Err: err.Error()} })
}

func (m *Mod) vote(_ peer.ID, c wire.VoteReq) wire.VoteResp {
	v, err := m.Decide(context.Background(), c)
	if err != nil {
		return wire.VoteResp{Err: err.Error()}
	}
	return wire.VoteResp{Vote: v}
}

// Decide checks c and returns this moderator's signed vote on it.
func (m *Mod) Decide(ctx context.Context, c proto.MsgCert) (*proto.Vote, error) {
	now := time.Now()
	if m.Now != nil {
		now = m.Now()
	}
	msg := c.Msg
	set, ok := m.Sets.Get(msg.Community, c.Epoch)
	if !ok {
		return nil, errors.New("unknown community or epoch")
	}
	if !set.IsMember(m.Key.Public().(ed25519.PublicKey)) {
		return nil, errors.New("not a moderator of this community")
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
	ok, err := m.Policy(ctx, msg)
	v := proto.SignModVote(m.Key, msg, c.Epoch, ok && err == nil)
	return &v, nil
}
