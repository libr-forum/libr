package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/libr/forum"
)

// Expiry drops pending items nobody decided within a week.
const Expiry = 7 * 24 * time.Hour

// PendingInfo is a post or report of the user's that is waiting for moderators, or
// (in OnFinalized) how it ended.
type PendingInfo struct {
	ID string `json:"id"`
	// Key is what the UI tracks the item by: the reported post's hash for a report,
	// ID for a post (which has no hash until it is published).
	Key       string `json:"key"`
	Kind      string `json:"kind"` // post | report
	Community string `json:"community"`
	Content   string `json:"content"`          // the post (own post, or the reported one)
	Reason    string `json:"reason,omitempty"` // reports
	Created   int64  `json:"created"`
	Approved  int    `json:"approved"`
	Rejected  int    `json:"rejected"`
	Waiting   int    `json:"waiting"` // moderators reviewing it
	Status    string `json:"status"`  // pending | posted | removed | rejected | expired
	Hash      string `json:"hash,omitempty"`
}

type record struct {
	ID      string        `json:"id"`
	P       forum.Pending `json:"pending"`
	Content string        `json:"content,omitempty"` // post text, to re-sign if no mod has it queued
	Created int64         `json:"created"`
}

func (r *record) info(status string) PendingInfo {
	ap, rj := r.P.Tally()
	key := r.ID
	if r.P.Kind == forum.KindReport {
		h := r.P.Cert.Hash()
		key = hex.EncodeToString(h[:])
	}
	return PendingInfo{ID: r.ID, Key: key, Kind: r.P.Kind, Community: r.P.Cert.Msg.Community, Content: string(r.P.Cert.Msg.Content),
		Reason: r.P.Reason, Created: r.Created, Approved: ap, Rejected: rj, Waiting: r.P.Waiting, Status: status}
}

// pendingStore keeps one JSON file per pending item.
type pendingStore struct {
	dir string
	mu  sync.Mutex
}

func openPending(dir string) (*pendingStore, error) {
	return &pendingStore{dir: dir}, os.MkdirAll(dir, 0o700)
}

func (s *pendingStore) add(p *forum.Pending, content string) (string, error) {
	id := p.Cert.Msg.ID()
	key := p.Kind + "-" + hex.EncodeToString(id[:12])
	if p.Kind == forum.KindReport {
		r := sha256.Sum256([]byte(p.Reason))
		key += "-" + hex.EncodeToString(r[:4])
	}
	return key, s.put(&record{ID: key, P: *p, Content: content, Created: time.Now().Unix()})
}

func (s *pendingStore) put(r *record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, r.ID+".json")
	if err := os.WriteFile(path+".tmp", b, 0o600); err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

func (s *pendingStore) del(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.Remove(filepath.Join(s.dir, id+".json"))
}

func (s *pendingStore) list() []*record {
	s.mu.Lock()
	defer s.mu.Unlock()
	names, _ := filepath.Glob(filepath.Join(s.dir, "*.json"))
	var out []*record
	for _, n := range names {
		b, err := os.ReadFile(n)
		if err != nil {
			continue
		}
		var r record
		if json.Unmarshal(b, &r) == nil && strings.HasSuffix(n, r.ID+".json") {
			out = append(out, &r)
		}
	}
	return out
}

// Pending lists the user's posts and reports waiting for moderators, newest first.
func (a *App) Pending() []PendingInfo {
	var out []PendingInfo
	for _, r := range a.store.list() {
		out = append(out, r.info("pending"))
	}
	return sortedInfos(out)
}

// Tick resumes every pending item once. It runs every Tick; adapters may also call it.
func (a *App) Tick(ctx context.Context) {
	a.tickMu.Lock()
	defer a.tickMu.Unlock()
	for _, r := range a.store.list() {
		if time.Since(time.Unix(r.Created, 0)) > Expiry {
			a.finish(r, "expired", "")
			continue
		}
		var posted *forum.Posted
		var removed *forum.Removed
		var err error
		if r.P.Kind == forum.KindPost && len(r.P.Votes) == 0 && r.P.Waiting == 0 {
			// No moderator has it queued, and they refuse an old timestamp: sign afresh.
			posted, err = a.client.Submit(ctx, r.P.Cert.Msg.Community, []byte(r.Content))
			var qe *forum.QuorumError
			if errors.As(err, &qe) {
				r.P = *qe.Pending
			}
		} else {
			posted, removed, err = a.client.Resume(ctx, &r.P)
		}
		var qe *forum.QuorumError
		switch {
		case errors.As(err, &qe):
			_ = a.store.put(r)
		case errors.Is(err, forum.ErrRejected):
			a.finish(r, "rejected", "")
		case err != nil:
			// network trouble; try again next tick
		case posted != nil:
			a.remember(posted.Cert)
			a.finish(r, "posted", hex.EncodeToString(posted.Hash[:]))
		case removed != nil:
			a.finish(r, "removed", "")
		}
	}
}

func (a *App) finish(r *record, status, hash string) {
	a.store.del(r.ID)
	if a.cfg.OnFinalized != nil {
		info := r.info(status)
		info.Hash = hash
		a.cfg.OnFinalized(info)
	}
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
