package wire

import (
	"crypto/ed25519"

	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// Protocol IDs. v1 used a single /chat/1.0.0 proxied by the relay; v2 streams go
// end to end and each operation has its own protocol.
const (
	ProtoStore    = protocol.ID("/libr/2/store")
	ProtoGet      = protocol.ID("/libr/2/get")
	ProtoIndexPut = protocol.ID("/libr/2/index/put")
	ProtoIndexGet = protocol.ID("/libr/2/index/get")
	ProtoVote     = protocol.ID("/libr/2/vote")   // served by moderators
	ProtoReport   = protocol.ID("/libr/2/report") // served by moderators
	ProtoRemove   = protocol.ID("/libr/2/remove")
)

// BucketSeconds is the width of one feed index bucket.
const BucketSeconds = 15 * 60

// MaxIndexHashes bounds one index/get reply.
const MaxIndexHashes = 1000

// Bucket returns the index bucket for a unix timestamp.
func Bucket(ts int64) int64 { return ts - ts%BucketSeconds }

// IndexKey is the DHT key under which a community's bucket is indexed.
func IndexKey(community string, bucket int64) proto.Hash {
	return proto.Sum(proto.Encode("libr/v2/index", []byte(community), proto.I64(bucket)))
}

// ReceiptResp answers store and index/put.
type ReceiptResp struct {
	Receipt *proto.StoreReceipt `json:"receipt,omitempty"`
	Err     string              `json:"err,omitempty"`
}

// GetReq asks for the cert stored under Hash.
type GetReq struct {
	Hash proto.Hash `json:"hash"`
}

// GetResp carries the cert, or Err "notfound".
type GetResp struct {
	Cert *proto.MsgCert `json:"cert,omitempty"`
	Err  string         `json:"err,omitempty"`
}

// IndexGetReq asks for the cert hashes in one community bucket.
type IndexGetReq struct {
	Community string `json:"community"`
	Bucket    int64  `json:"bucket"`
}

// IndexGetResp lists cert hashes and the bucket's tombstones, at most MaxIndexHashes
// of each. Readers drop any message a verified tombstone removes.
type IndexGetResp struct {
	Hashes []proto.Hash      `json:"hashes,omitempty"`
	Tombs  []proto.Tombstone `json:"tombs,omitempty"`
	Err    string            `json:"err,omitempty"`
}

// ReportReq asks a moderator to vote on removing Cert's message. The reporter signs
// proto.ReportBytes(community, Epoch, msgID, Reason) so reports are attributable and
// rate-limited per reporter.
type ReportReq struct {
	Cert     proto.MsgCert     `json:"cert"`
	Epoch    uint64            `json:"epoch"` // epoch of the RepCert being collected
	Reason   string            `json:"reason"`
	Reporter ed25519.PublicKey `json:"reporter"`
	Sig      []byte            `json:"sig"`
}

// RemoveReq asks a DB node to tombstone Msg. Msg is sent so the node can check the
// proof (author signature, community) without holding a copy of the post.
type RemoveReq struct {
	Msg  proto.Msg       `json:"msg"`
	Tomb proto.Tombstone `json:"tomb"`
}

// RemoveResp carries the node's signed ack.
type RemoveResp struct {
	Ack *proto.RemoveAck `json:"ack,omitempty"`
	Err string           `json:"err,omitempty"`
}

// VoteReq asks a moderator to decide on an author-signed message (a MsgCert with no
// votes, Epoch set to the epoch the author is collecting for).
type VoteReq = proto.MsgCert

// VoteResp carries the moderator's signed decision; a rejection is a signed vote
// with Approve false. Err is set when the request itself was refused.
type VoteResp struct {
	Vote *proto.Vote `json:"vote,omitempty"`
	Err  string      `json:"err,omitempty"`
}

// Error strings carried in Err fields.
const (
	ErrNotFound = "notfound"
	ErrInternal = "internal"
	ErrRemoved  = "removed" // the message has a tombstone; no new copies are accepted
)
