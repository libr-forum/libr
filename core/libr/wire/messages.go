package wire

import (
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
	ProtoVote     = protocol.ID("/libr/2/vote") // served by moderators
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

// IndexGetResp lists cert hashes, at most MaxIndexHashes.
type IndexGetResp struct {
	Hashes []proto.Hash `json:"hashes,omitempty"`
	Err    string       `json:"err,omitempty"`
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
)
