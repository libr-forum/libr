// Package dbnode serves LIBR v2 storage: it accepts only certificates that verify
// against the known moderator sets (fail closed), stores them by hash, indexes them per
// community time bucket, and answers with a receipt signed by its own key.
package dbnode

import (
	"crypto/ed25519"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// Store persists certificates and the feed index. Put and AddIndex are idempotent.
type Store interface {
	Put(h proto.Hash, c proto.MsgCert) error
	Get(h proto.Hash) (*proto.MsgCert, error) // nil, nil when absent
	AddIndex(community string, bucket int64, h proto.Hash) error
	Index(community string, bucket int64, limit int) ([]proto.Hash, error)
}

// Node is a storage node's protocol logic.
type Node struct {
	Key   ed25519.PrivateKey
	Sets  proto.ModSets
	Store Store
	Now   func() time.Time // nil means time.Now
}

// Per-peer request budget: generous for a client posting and reading feeds, small
// enough that one peer cannot monopolise a node.
const (
	ratePerSecond = 20
	rateBurst     = 100
)

// Register installs the v2 storage protocols on h.
func (n *Node) Register(h host.Host) {
	lim := wire.NewLimiter(ratePerSecond, rateBurst)
	rejectReceipt := func(err error) wire.ReceiptResp { return wire.ReceiptResp{Err: err.Error()} }
	wire.Handle(h, wire.ProtoStore, lim, n.store, rejectReceipt)
	wire.Handle(h, wire.ProtoIndexPut, lim, n.indexPut, rejectReceipt)
	wire.Handle(h, wire.ProtoGet, lim, n.get, func(err error) wire.GetResp { return wire.GetResp{Err: err.Error()} })
	wire.Handle(h, wire.ProtoIndexGet, lim, n.indexGet, func(err error) wire.IndexGetResp { return wire.IndexGetResp{Err: err.Error()} })
}

func (n *Node) now() time.Time {
	if n.Now != nil {
		return n.Now()
	}
	return time.Now()
}

func (n *Node) store(_ peer.ID, c proto.MsgCert) wire.ReceiptResp {
	if err := proto.VerifyMsgCert(n.Sets, c, n.now()); err != nil {
		return wire.ReceiptResp{Err: "invalid: " + err.Error()}
	}
	if err := n.Store.Put(c.Hash(), c); err != nil {
		return wire.ReceiptResp{Err: wire.ErrInternal}
	}
	return n.receipt(c)
}

// indexPut takes the full cert, not just its hash, so the index only ever lists
// certificates this node has verified.
func (n *Node) indexPut(_ peer.ID, c proto.MsgCert) wire.ReceiptResp {
	if err := proto.VerifyMsgCert(n.Sets, c, n.now()); err != nil {
		return wire.ReceiptResp{Err: "invalid: " + err.Error()}
	}
	if err := n.Store.AddIndex(c.Msg.Community, wire.Bucket(c.Msg.Ts), c.Hash()); err != nil {
		return wire.ReceiptResp{Err: wire.ErrInternal}
	}
	return n.receipt(c)
}

func (n *Node) receipt(c proto.MsgCert) wire.ReceiptResp {
	r := proto.SignReceipt(n.Key, c, n.now().Unix())
	return wire.ReceiptResp{Receipt: &r}
}

func (n *Node) get(_ peer.ID, req wire.GetReq) wire.GetResp {
	c, err := n.Store.Get(req.Hash)
	switch {
	case err != nil:
		return wire.GetResp{Err: wire.ErrInternal}
	case c == nil:
		return wire.GetResp{Err: wire.ErrNotFound}
	}
	return wire.GetResp{Cert: c}
}

func (n *Node) indexGet(_ peer.ID, req wire.IndexGetReq) wire.IndexGetResp {
	hs, err := n.Store.Index(req.Community, req.Bucket, wire.MaxIndexHashes)
	if err != nil {
		return wire.IndexGetResp{Err: wire.ErrInternal}
	}
	return wire.IndexGetResp{Hashes: hs}
}
