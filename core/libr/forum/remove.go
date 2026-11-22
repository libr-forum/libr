package forum

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

// Removed is the outcome of Delete or Report.
type Removed struct {
	Tomb proto.Tombstone
	Acks []proto.RemoveAck // from the cert's replicas, at return (>= StoredQuorum)
	// IndexAcks come from the message's feed bucket nodes, which also refuse any later
	// certificate of the message and hand the tombstone to every feed reader.
	IndexAcks []proto.RemoveAck
	Done      <-chan struct{}
}

// Delete withdraws the client's own post.
func (c *Client) Delete(ctx context.Context, cert proto.MsgCert) (*Removed, error) {
	if !bytes.Equal(cert.Msg.Author, c.Key.Public().(ed25519.PublicKey)) {
		return nil, errors.New("only the author can delete a post")
	}
	d := proto.SignAuthorDelete(c.Key, cert.Msg)
	return c.remove(ctx, cert.Msg, cert.Hash(), proto.Tombstone{MsgID: d.MsgID, Del: &d})
}

// Report asks the community's moderators to remove cert's message for reason. With a
// removal quorum it tombstones the message and returns the RepCert; ErrRejected if
// the moderators keep it.
func (c *Client) Report(ctx context.Context, cert proto.MsgCert, reason string) (*Removed, error) {
	m := cert.Msg
	set, ok := c.Sets.Head(m.Community)
	if !ok {
		return nil, fmt.Errorf("unknown community %q", m.Community)
	}
	rc := proto.RepCert{MsgID: m.ID(), Community: m.Community, Epoch: set.Epoch, Reason: reason}
	req := wire.ReportReq{Cert: cert, Epoch: set.Epoch, Reason: reason, Reporter: c.Key.Public().(ed25519.PublicKey),
		Sig: ed25519.Sign(c.Key, proto.ReportBytes(m.Community, set.Epoch, rc.MsgID, reason))}
	ballot := rc // late answers still read it after quorum returns; rc gets the votes
	votes, err := c.quorum(ctx, set, func(ctx context.Context, mod proto.ModEntry) (*proto.Vote, error) {
		p, err := peer.Decode(mod.PeerID)
		if err != nil {
			return nil, err
		}
		var resp wire.VoteResp
		err = c.call(ctx, "report", p, wire.ProtoReport, req, &resp, func() error {
			if resp.Err != "" {
				return errors.New(resp.Err)
			}
			if resp.Vote == nil || !proto.VerifyRepVote(*resp.Vote, mod.PubKey, ballot) {
				return errors.New("invalid vote")
			}
			return nil
		})
		return resp.Vote, err
	})
	if err != nil {
		return nil, err
	}
	rc.Votes = votes
	if err := proto.VerifyRepCert(c.Sets, rc); err != nil {
		return nil, fmt.Errorf("assembled removal certificate does not verify: %w", err)
	}
	return c.remove(ctx, m, cert.Hash(), proto.Tombstone{MsgID: rc.MsgID, Rep: &rc})
}

// remove sends the tombstone to the nodes holding the cert and to the message's feed
// bucket nodes, each needing StoredQuorum signed acks.
func (c *Client) remove(ctx context.Context, m proto.Msg, certHash proto.Hash, t proto.Tombstone) (*Removed, error) {
	req := wire.RemoveReq{Msg: m, Tomb: t}
	send := func(op string, key proto.Hash, acks *[]proto.RemoveAck) (<-chan struct{}, error) {
		var mu sync.Mutex
		var got []proto.RemoveAck
		done, err := c.spread(ctx, op, key, func(ctx context.Context, p peer.ID) bool {
			var resp wire.RemoveResp
			err := c.call(ctx, op, p, wire.ProtoRemove, req, &resp, func() error {
				if resp.Err != "" {
					return errors.New(resp.Err)
				}
				return checkAck(resp.Ack, t.MsgID, p)
			})
			if err == nil {
				mu.Lock()
				got = append(got, *resp.Ack)
				mu.Unlock()
			}
			return err == nil
		})
		mu.Lock()
		*acks = append([]proto.RemoveAck(nil), got...)
		mu.Unlock()
		return done, err
	}
	r := &Removed{Tomb: t}
	var err error
	r.Done, err = both(
		func() (<-chan struct{}, error) { return send("remove", certHash, &r.Acks) },
		func() (<-chan struct{}, error) { return send("remove_index", indexKeyOf(m), &r.IndexAcks) },
	)
	return r, err
}

// checkAck accepts only an ack for msgID signed by the peer we talked to.
func checkAck(a *proto.RemoveAck, msgID proto.Hash, from peer.ID) error {
	if a == nil || !a.Verify() || a.MsgID != msgID {
		return errors.New("invalid ack")
	}
	if id, err := proto.PeerIDFor(a.Node); err != nil || id != from.String() {
		return errors.New("ack signed by another node")
	}
	return nil
}
