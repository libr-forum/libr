package proto

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"
)

const (
	// MaxContent bounds a message body (images included).
	MaxContent = 1 << 20
	// MaxClockSkew is how far in the future a message timestamp may be.
	MaxClockSkew = 5 * time.Minute
	// MaxReason bounds a removal reason.
	MaxReason = 1024
)

// ModSets is what verification needs from a Registry.
type ModSets interface {
	Get(community string, epoch uint64) (*ModSet, bool)
}

// VerifyMsgCert accepts c only if its author signature is valid and at least
// Threshold distinct members of the community's moderator set for c.Epoch approved
// it. Anything unknown or malformed is rejected.
func VerifyMsgCert(sets ModSets, c MsgCert, now time.Time) error {
	m := c.Msg
	if len(m.Content) == 0 || len(m.Content) > MaxContent {
		return fmt.Errorf("content length %d out of range", len(m.Content))
	}
	if time.Unix(m.Ts, 0).After(now.Add(MaxClockSkew)) {
		return errors.New("timestamp in the future")
	}
	if !verify(m.Author, m.Bytes(), c.AuthorSig) {
		return errors.New("bad author signature")
	}
	set, ok := sets.Get(m.Community, c.Epoch)
	if !ok {
		return fmt.Errorf("unknown community %q epoch %d", m.Community, c.Epoch)
	}
	return quorum(set, c.Votes, ModVoteBytes(m.Community, c.Epoch, m.ID(), true))
}

// VerifyRepCert accepts a removal only with a Threshold quorum of the community's
// moderators approving it.
func VerifyRepCert(sets ModSets, r RepCert) error {
	if len(r.Reason) > MaxReason {
		return errors.New("reason too long")
	}
	set, ok := sets.Get(r.Community, r.Epoch)
	if !ok {
		return fmt.Errorf("unknown community %q epoch %d", r.Community, r.Epoch)
	}
	return quorum(set, r.Votes, RepVoteBytes(r.Community, r.Epoch, r.MsgID, r.Reason, true))
}

// VerifyAuthorDelete accepts d only for msg's own ID, signed by msg's author.
func VerifyAuthorDelete(d AuthorDelete, msg Msg) error {
	if d.MsgID != msg.ID() {
		return errors.New("delete is for a different message")
	}
	if !verify(msg.Author, AuthorDeleteBytes(d.MsgID), d.Sig) {
		return errors.New("not signed by the author")
	}
	return nil
}

// quorum requires every vote to be an approval from a distinct member with a valid
// signature over the approval payload, and at least set.Threshold of them. Rejections do not
// belong in a certificate; votes beyond the member count are refused outright.
func quorum(set *ModSet, votes []Vote, payload []byte) error {
	if len(votes) > len(set.Mods) {
		return fmt.Errorf("%d votes for %d moderators", len(votes), len(set.Mods))
	}
	seen := make(map[string]bool, len(votes))
	for _, v := range votes {
		if !v.Approve {
			return errors.New("certificate contains a non-approving vote")
		}
		k := string(v.Mod)
		if len(v.Mod) != ed25519.PublicKeySize || !set.IsMember(v.Mod) {
			return errors.New("vote from a non-member")
		}
		if seen[k] {
			return errors.New("duplicate vote")
		}
		seen[k] = true
		if !verify(v.Mod, payload, v.Sig) {
			return errors.New("bad vote signature")
		}
	}
	if uint64(len(seen)) < set.Threshold {
		return fmt.Errorf("%d approvals, need %d", len(seen), set.Threshold)
	}
	return nil
}
