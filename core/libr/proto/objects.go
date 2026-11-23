package proto

import (
	"bytes"
	"crypto/ed25519"
	"sort"
)

// Msg is a post as written by its author.
type Msg struct {
	Community string            `json:"community"`
	Author    ed25519.PublicKey `json:"author"`
	Ts        int64             `json:"ts"` // unix seconds
	Content   []byte            `json:"content"`
}

// Bytes is what the author signs.
func (m Msg) Bytes() []byte {
	return Encode(DomainMsg, []byte(m.Community), m.Author, I64(m.Ts), m.Content)
}

// ID identifies the message independent of which moderators approved it. Removals
// (tombstones) are keyed by ID, so a second valid cert for the same message is
// removed too.
func (m Msg) ID() Hash { return Sum(m.Bytes()) }

// Vote is one moderator's decision on a message or a removal request.
type Vote struct {
	Mod     ed25519.PublicKey `json:"mod"`
	Approve bool              `json:"approve"`
	Sig     []byte            `json:"sig"`
}

// ModVoteBytes is what a moderator signs to decide on publishing msgID. Binding the
// community and epoch stops a vote being reused elsewhere.
func ModVoteBytes(community string, epoch uint64, msgID Hash, approve bool) []byte {
	return Encode(DomainModVote, []byte(community), U64(epoch), msgID[:], Bool(approve))
}

// SignModVote returns priv's vote on m in epoch.
func SignModVote(priv ed25519.PrivateKey, m Msg, epoch uint64, approve bool) Vote {
	return Vote{Mod: pub(priv), Approve: approve, Sig: ed25519.Sign(priv, ModVoteBytes(m.Community, epoch, m.ID(), approve))}
}

// MsgCert is a message with its author signature and a quorum of approvals from the
// community's moderator set for Epoch.
type MsgCert struct {
	Msg       Msg    `json:"msg"`
	AuthorSig []byte `json:"author_sig"`
	Epoch     uint64 `json:"epoch"`
	Votes     []Vote `json:"votes"`
}

// SignMsg returns m signed by its author (an unvoted cert, sent to moderators).
func SignMsg(priv ed25519.PrivateKey, m Msg) MsgCert {
	m.Author = pub(priv)
	return MsgCert{Msg: m, AuthorSig: ed25519.Sign(priv, m.Bytes())}
}

// Hash is the cert's storage key. Votes are hashed in key order, so the order
// moderators answered in does not matter.
func (c MsgCert) Hash() Hash {
	fields := [][]byte{c.Msg.Bytes(), c.AuthorSig, U64(c.Epoch)}
	for _, v := range sortedVotes(c.Votes) {
		fields = append(fields, v.Mod, Bool(v.Approve), v.Sig)
	}
	return Sum(Encode(DomainMsgCert, fields...))
}

// RepCert removes a message: a quorum of the community's moderators voting to remove
// MsgID for Reason.
type RepCert struct {
	MsgID     Hash   `json:"msg_id"`
	Community string `json:"community"`
	Epoch     uint64 `json:"epoch"`
	Reason    string `json:"reason"`
	Votes     []Vote `json:"votes"`
}

// RepVoteBytes is what a moderator signs to decide on removing msgID.
func RepVoteBytes(community string, epoch uint64, msgID Hash, reason string, approve bool) []byte {
	return Encode(DomainRepVote, []byte(community), U64(epoch), msgID[:], []byte(reason), Bool(approve))
}

// SignRepVote returns priv's vote on removing r.MsgID.
func SignRepVote(priv ed25519.PrivateKey, r RepCert, approve bool) Vote {
	return Vote{Mod: pub(priv), Approve: approve, Sig: ed25519.Sign(priv, RepVoteBytes(r.Community, r.Epoch, r.MsgID, r.Reason, approve))}
}

// AuthorDelete is the author withdrawing their own message.
type AuthorDelete struct {
	MsgID Hash   `json:"msg_id"`
	Sig   []byte `json:"sig"`
}

// AuthorDeleteBytes is what the author signs to withdraw msgID.
func AuthorDeleteBytes(msgID Hash) []byte { return Encode(DomainAuthorDelete, msgID[:]) }

// SignAuthorDelete withdraws m; priv must be m's author.
func SignAuthorDelete(priv ed25519.PrivateKey, m Msg) AuthorDelete {
	id := m.ID()
	return AuthorDelete{MsgID: id, Sig: ed25519.Sign(priv, AuthorDeleteBytes(id))}
}

// StoreReceipt is a DB node's signed statement that it stored CertHash. Tagged with
// the community so storage work can later count per community.
type StoreReceipt struct {
	CertHash  Hash              `json:"cert_hash"`
	Community string            `json:"community"`
	Node      ed25519.PublicKey `json:"node"`
	Ts        int64             `json:"ts"`
	Sig       []byte            `json:"sig"`
}

func (r StoreReceipt) Bytes() []byte {
	return Encode(DomainReceipt, r.CertHash[:], []byte(r.Community), r.Node, I64(r.Ts))
}

// SignReceipt returns priv's receipt for storing c at ts.
func SignReceipt(priv ed25519.PrivateKey, c MsgCert, ts int64) StoreReceipt {
	r := StoreReceipt{CertHash: c.Hash(), Community: c.Msg.Community, Node: pub(priv), Ts: ts}
	r.Sig = ed25519.Sign(priv, r.Bytes())
	return r
}

// Verify checks the receipt is signed by its node.
func (r StoreReceipt) Verify() bool { return verify(r.Node, r.Bytes(), r.Sig) }

func pub(priv ed25519.PrivateKey) ed25519.PublicKey { return priv.Public().(ed25519.PublicKey) }

func verify(k ed25519.PublicKey, msg, sig []byte) bool {
	return len(k) == ed25519.PublicKeySize && len(sig) == ed25519.SignatureSize && ed25519.Verify(k, msg, sig)
}

func sortedVotes(vs []Vote) []Vote {
	out := append([]Vote(nil), vs...)
	sort.Slice(out, func(i, j int) bool {
		if c := bytes.Compare(out[i].Mod, out[j].Mod); c != 0 {
			return c < 0
		}
		return bytes.Compare(out[i].Sig, out[j].Sig) < 0
	})
	return out
}
