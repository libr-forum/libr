package moderator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// ErrDefer, returned by a Policy or ReportPolicy, sends the item to the human review
// queue instead of deciding now.
var ErrDefer = errors.New("deferred to a human moderator")

// ErrPending answers a request whose item is still waiting in the queue.
var ErrPending = errors.New("pending human review")

const (
	maxQueued = 1000 // undecided items; beyond this, deferrals are refused

	keepFor = 7 * 24 * time.Hour // decided or not, items are dropped after this
)

// Item is one post or report waiting for (or decided by) a human moderator.
type Item struct {
	Key      string        `json:"key"`
	Kind     string        `json:"kind"` // "post" | "report"
	Cert     proto.MsgCert `json:"cert"`
	Epoch    uint64        `json:"epoch"`
	Reason   string        `json:"reason,omitempty"`
	Received int64         `json:"received"`
	Vote     *proto.Vote   `json:"vote,omitempty"` // set once decided
}

// PostKey and ReportKey identify queue items; the author or reporter asking again
// maps to the same item.
func PostKey(m proto.Msg, epoch uint64) string {
	id := m.ID()
	return fmt.Sprintf("post/%s/%d", hex.EncodeToString(id[:]), epoch)
}

// ReportKey identifies a moderator's one decision on removing a post: every report
// of the post, whatever its reason, maps to it.
func ReportKey(m proto.Msg, epoch uint64) string {
	id := m.ID()
	return fmt.Sprintf("report/%s/%d", hex.EncodeToString(id[:]), epoch)
}

// ContentKey identifies a post's content in a community, independent of author and
// timestamp: a moderator that rejected it rejects identical content again.
func ContentKey(m proto.Msg, epoch uint64) string {
	h := sha256.Sum256(proto.Encode("libr/v2/content", []byte(m.Community), m.Content))
	return fmt.Sprintf("content/%s/%d", hex.EncodeToString(h[:]), epoch)
}

// Queue holds deferred items, persisted as one JSON file (empty path: memory only).
type Queue struct {
	mu    sync.Mutex
	path  string
	items map[string]*Item
	now   func() time.Time
}

// NewQueue loads the queue at path, if any.
func NewQueue(path string) (*Queue, error) {
	q := &Queue{path: path, items: map[string]*Item{}, now: time.Now}
	if path == "" {
		return q, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return q, nil
	}
	if err != nil {
		return nil, err
	}
	var list []*Item
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("moderation queue %s: %w", path, err)
	}
	for _, it := range list {
		q.items[it.Key] = it
	}
	return q, nil
}

// Items returns every item, undecided first, oldest first.
func (q *Queue) Items() []Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.prune()
	out := make([]Item, 0, len(q.items))
	for _, it := range q.items {
		out = append(out, *it)
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Vote == nil) != (out[j].Vote == nil) {
			return out[i].Vote == nil
		}
		return out[i].Received < out[j].Received
	})
	return out
}

func (q *Queue) get(key string) (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	it, ok := q.items[key]
	if !ok {
		return Item{}, false
	}
	return *it, true
}

func (q *Queue) add(it Item) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.items[it.Key]; ok {
		return nil
	}
	q.prune()
	undecided := 0
	for _, x := range q.items {
		if x.Vote == nil {
			undecided++
		}
	}
	if undecided >= maxQueued {
		return ErrBusy
	}
	it.Received = q.now().Unix()
	q.items[it.Key] = &it
	return q.save()
}

func (q *Queue) decide(key string, v proto.Vote) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	it, ok := q.items[key]
	if !ok {
		return fmt.Errorf("no queued item %q", key)
	}
	if it.Vote != nil {
		return errors.New("already decided")
	}
	it.Vote = &v
	return q.save()
}

func (q *Queue) prune() {
	cut := q.now().Add(-keepFor).Unix()
	for k, it := range q.items {
		if it.Received < cut {
			delete(q.items, k)
		}
	}
}

// save writes the queue atomically. Caller holds mu.
func (q *Queue) save() error {
	if q.path == "" {
		return nil
	}
	list := make([]*Item, 0, len(q.items))
	for _, it := range q.items {
		list = append(list, it)
	}
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	tmp := q.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, q.path)
}
