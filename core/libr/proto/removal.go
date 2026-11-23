package proto

import (
	"crypto/ed25519"
	"errors"
)

// Tombstone removes a message (every certificate of it, since it is keyed by MsgID).
// Exactly one proof is set: a moderator quorum (Rep) or the author (Del).
type Tombstone struct {
	MsgID Hash          `json:"msg_id"`
	Rep   *RepCert      `json:"rep,omitempty"`
	Del   *AuthorDelete `json:"del,omitempty"`
}

// Kind is "mod" or "author".
func (t Tombstone) Kind() string {
	if t.Rep != nil {
		return "mod"
	}
	return "author"
}

// VerifyTombstone checks t removes m and its proof holds. A RepCert only removes
// messages of its own community, so one community's moderators cannot remove another's
// posts.
func VerifyTombstone(sets ModSets, t Tombstone, m Msg) error {
	if t.MsgID != m.ID() {
		return errors.New("tombstone is for a different message")
	}
	switch {
	case t.Rep != nil && t.Del == nil:
		if t.Rep.MsgID != t.MsgID || t.Rep.Community != m.Community {
			return errors.New("removal certificate is for another message or community")
		}
		return VerifyRepCert(sets, *t.Rep)
	case t.Del != nil && t.Rep == nil:
		return VerifyAuthorDelete(*t.Del, m)
	}
	return errors.New("tombstone needs exactly one proof")
}

// ReportBytes is what a reporter signs to ask community's moderators (of epoch) to
// remove msgID for reason.
func ReportBytes(community string, epoch uint64, msgID Hash, reason string) []byte {
	return Encode(DomainReport, []byte(community), U64(epoch), msgID[:], []byte(reason))
}

// RemoveAck is a DB node's signed statement that it holds the tombstone for MsgID
// and has deleted the message.
type RemoveAck struct {
	MsgID Hash              `json:"msg_id"`
	Node  ed25519.PublicKey `json:"node"`
	Ts    int64             `json:"ts"`
	Sig   []byte            `json:"sig"`
}

func (a RemoveAck) Bytes() []byte {
	return Encode(DomainRemoveAck, a.MsgID[:], a.Node, I64(a.Ts))
}

// SignRemoveAck returns priv's ack for removing msgID at ts.
func SignRemoveAck(priv ed25519.PrivateKey, msgID Hash, ts int64) RemoveAck {
	a := RemoveAck{MsgID: msgID, Node: pub(priv), Ts: ts}
	a.Sig = ed25519.Sign(priv, a.Bytes())
	return a
}

// Verify checks the ack is signed by its node.
func (a RemoveAck) Verify() bool { return verify(a.Node, a.Bytes(), a.Sig) }

// VerifyRepVote checks v is mod's signed decision on removing r.MsgID.
func VerifyRepVote(v Vote, mod ed25519.PublicKey, r RepCert) bool {
	return string(v.Mod) == string(mod) && verify(mod, RepVoteBytes(r.Community, r.Epoch, r.MsgID, r.Reason, v.Approve), v.Sig)
}
