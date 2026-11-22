package dbnode

import (
	"fmt"
	"strings"
	"testing"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

func repCert(c proto.MsgCert, voters ...int) *proto.RepCert {
	r := proto.RepCert{MsgID: c.Msg.ID(), Community: c.Msg.Community, Reason: "spam"}
	for _, i := range voters {
		r.Votes = append(r.Votes, proto.SignRepVote(key(fmt.Sprint("mod", i)), r, true))
	}
	return &r
}

// TestRemove: invalid removals are refused and change nothing; a valid one deletes the
// message, lists its tombstone in the bucket, and refuses any later copy of it.
func TestRemove(t *testing.T) {
	reg, mods := community(t)
	r := newRig(t, reg)
	author := key("author")
	post := cert(author, "to be removed", now, mods[0], mods[1], mods[2])
	other := cert(author, "stays", now, mods[0], mods[1], mods[2])
	for _, c := range []proto.MsgCert{post, other} {
		var rr wire.ReceiptResp
		r.call(t, string(wire.ProtoStore), c, &rr)
		r.call(t, string(wire.ProtoIndexPut), c, &rr)
	}

	strangerDel := proto.SignAuthorDelete(key("stranger"), post.Msg)
	wrongComm := post.Msg
	wrongComm.Community = "nope"
	for name, req := range map[string]wire.RemoveReq{
		"rep cert with 2f votes":      {Msg: post.Msg, Tomb: proto.Tombstone{MsgID: post.Msg.ID(), Rep: repCert(post, 0, 1)}},
		"delete by someone else":      {Msg: post.Msg, Tomb: proto.Tombstone{MsgID: post.Msg.ID(), Del: &strangerDel}},
		"tombstone for other message": {Msg: other.Msg, Tomb: proto.Tombstone{MsgID: post.Msg.ID(), Rep: repCert(post, 0, 1, 2)}},
		"unknown community":           {Msg: wrongComm, Tomb: proto.Tombstone{MsgID: wrongComm.ID(), Rep: repCert(post, 0, 1, 2)}},
	} {
		var resp wire.RemoveResp
		r.call(t, string(wire.ProtoRemove), req, &resp)
		if resp.Ack != nil || !strings.HasPrefix(resp.Err, "invalid") {
			t.Errorf("%s: %+v", name, resp)
		}
	}
	if got, _ := r.store.Get(post.Hash()); got == nil {
		t.Fatal("refused removal deleted the post")
	}

	var resp wire.RemoveResp
	r.call(t, string(wire.ProtoRemove), wire.RemoveReq{Msg: post.Msg, Tomb: proto.Tombstone{MsgID: post.Msg.ID(), Rep: repCert(post, 0, 1, 2)}}, &resp)
	if resp.Ack == nil || !resp.Ack.Verify() || resp.Ack.MsgID != post.Msg.ID() {
		t.Fatalf("valid removal: %+v", resp)
	}
	if id, _ := proto.PeerIDFor(resp.Ack.Node); id != r.db.String() {
		t.Fatal("ack not signed by the node")
	}
	var gr wire.GetResp
	r.call(t, string(wire.ProtoGet), wire.GetReq{Hash: post.Hash()}, &gr)
	if gr.Err != wire.ErrNotFound {
		t.Fatalf("removed post still served: %+v", gr)
	}
	var ir wire.IndexGetResp
	r.call(t, string(wire.ProtoIndexGet), wire.IndexGetReq{Community: "general", Bucket: wire.Bucket(now)}, &ir)
	if len(ir.Hashes) != 1 || ir.Hashes[0] != other.Hash() || len(ir.Tombs) != 1 || ir.Tombs[0].MsgID != post.Msg.ID() {
		t.Fatalf("index after removal: %+v", ir)
	}

	// The same message re-approved by a different quorum is a different certificate.
	again := proto.MsgCert{Msg: post.Msg, AuthorSig: post.AuthorSig}
	for _, m := range mods[1:] {
		again.Votes = append(again.Votes, proto.SignModVote(m, again.Msg, 0, true))
	}
	if again.Hash() == post.Hash() {
		t.Fatal("test needs a second, distinct certificate")
	}
	for _, p := range []string{string(wire.ProtoStore), string(wire.ProtoIndexPut)} {
		var rr wire.ReceiptResp
		r.call(t, p, again, &rr)
		if rr.Receipt != nil || rr.Err != wire.ErrRemoved {
			t.Errorf("second cert via %s: %+v", p, rr)
		}
	}
}
