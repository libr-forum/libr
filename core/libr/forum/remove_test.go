package forum

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

func ctx90() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 90*time.Second)
}

// TestDeleteRemovesEverywhere: the author's delete reaches a quorum of the cert's
// replicas and of its bucket's index nodes; the post disappears from the feed and
// from every replica once background removal finishes.
func TestDeleteRemovesEverywhere(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 12 hosts")
	}
	w := newWorld(t, 1, 6, moderator.ApproveAll)
	gone, err := w.submit("delete me")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.submit("keep me"); err != nil {
		t.Fatal(err)
	}
	<-gone.Done
	ctx, cancel := ctx90()
	defer cancel()
	r, err := w.client.Delete(ctx, gone.Cert)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Acks) < StoredQuorum || len(r.IndexAcks) < StoredQuorum || r.Tomb.Kind() != "author" {
		t.Fatalf("acks %d, index acks %d, kind %s", len(r.Acks), len(r.IndexAcks), r.Tomb.Kind())
	}
	<-r.Done
	if got := strings.Join(w.feed(), "|"); got != "keep me" {
		t.Fatalf("feed after delete = %q", got)
	}
	for i, st := range w.stores {
		if c, _ := st.Get(gone.Hash); c != nil {
			t.Errorf("db%d still holds the deleted post", i)
		}
	}
}

// TestForgedDeleteRefused: a delete signed by anyone but the author gets no acks.
func TestForgedDeleteRefused(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	w := newWorld(t, 1, 4, moderator.ApproveAll)
	p, err := w.submit("mine")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := ctx90()
	defer cancel()
	if _, err := w.client.Delete(ctx, proto.MsgCert{Msg: proto.Msg{Author: pub(key("x"))}}); err == nil {
		t.Fatal("client deleted someone else's post")
	}
	forged := proto.SignAuthorDelete(key("stranger"), p.Cert.Msg)
	_, err = w.client.remove(ctx, p.Cert.Msg, p.Hash, proto.Tombstone{MsgID: forged.MsgID, Del: &forged})
	if !errors.Is(err, ErrNotStored) {
		t.Fatalf("forged delete: err = %v, want ErrNotStored", err)
	}
	if got := w.feed(); len(got) != 1 {
		t.Fatalf("feed after forged delete = %q", got)
	}
}

// TestReport: moderators remove a reported post by quorum, or keep it.
func TestReport(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	removeMe := func(_ context.Context, m proto.Msg, _ string) (bool, error) {
		return strings.Contains(string(m.Content), "remove-me"), nil
	}
	w := newWorld(t, 1, 4, moderator.ApproveAll, withReport(removeMe))
	bad, err := w.submit("please remove-me")
	if err != nil {
		t.Fatal(err)
	}
	fine, err := w.submit("perfectly fine")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := ctx90()
	defer cancel()
	if _, err := w.client.Report(ctx, fine.Cert, "dislike"); !errors.Is(err, ErrRejected) {
		t.Fatalf("report of a fine post: err = %v, want ErrRejected", err)
	}
	r, err := w.client.Report(ctx, bad.Cert, "breaks rules")
	if err != nil {
		t.Fatal(err)
	}
	if r.Tomb.Kind() != "mod" || len(r.Tomb.Rep.Votes) < 3 {
		t.Fatalf("tombstone %+v", r.Tomb)
	}
	<-r.Done
	if got := strings.Join(w.feed(), "|"); got != "perfectly fine" {
		t.Fatalf("feed after removal = %q", got)
	}
}

// TestSecondCertSuppressed: after removal, the same message re-approved under a
// different quorum (a different certificate hash) cannot be indexed again, and readers
// still hide it.
func TestSecondCertSuppressed(t *testing.T) {
	if testing.Short() {
		t.Skip("starts 10 hosts")
	}
	w := newWorld(t, 1, 4, moderator.ApproveAll)
	p, err := w.submit("removed once")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := ctx90()
	defer cancel()
	r, err := w.client.Delete(ctx, p.Cert)
	if err != nil {
		t.Fatal(err)
	}
	<-r.Done

	again := proto.MsgCert{Msg: p.Cert.Msg, AuthorSig: p.Cert.AuthorSig}
	for _, m := range w.modKeys {
		again.Votes = append(again.Votes, proto.SignModVote(m, again.Msg, 0, true))
	}
	if again.Hash() == p.Hash {
		t.Fatal("need a distinct certificate")
	}
	_, done, err := w.client.replicate(ctx, "index_put", wire.ProtoIndexPut, indexKeyOf(again.Msg), again)
	if !errors.Is(err, ErrNotStored) {
		t.Fatalf("re-index of removed message: err = %v", err)
	}
	<-done
	_, done, _ = w.client.replicate(ctx, "store", wire.ProtoStore, again.Hash(), again)
	<-done
	if got := w.feed(); len(got) != 0 {
		t.Fatalf("feed shows the removed message again: %q", got)
	}
}
