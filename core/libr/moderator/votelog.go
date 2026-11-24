package moderator

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// votes remembers every vote this moderator signed, so a repeated request gets the
// same vote: an author cannot re-ask until a nondeterministic policy approves, which
// is what makes "f+1 honest rejections are final" hold. A post entry is kept until no
// new vote on it could be signed (its timestamp leaves the clock-skew window); a
// report entry for keepFor. Persisted as JSON lines (empty path: memory only), so a
// restart cannot sign twice either.
type votes struct {
	mu   sync.Mutex
	path string
	m    map[string]voteRec
	f    *os.File
}

type voteRec struct {
	Key   string     `json:"key"`
	Until int64      `json:"until"`
	Vote  proto.Vote `json:"vote"`
}

func openVotes(path string, now time.Time) (*votes, error) {
	v := &votes{path: path, m: map[string]voteRec{}}
	if path == "" {
		return v, nil
	}
	if f, err := os.Open(path); err == nil {
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for sc.Scan() {
			var r voteRec
			if json.Unmarshal(sc.Bytes(), &r) == nil && r.Until >= now.Unix() {
				v.m[r.Key] = r
			}
		}
		f.Close()
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	// Rewrite without expired entries, then append from here on.
	tmp := path + ".tmp"
	w, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(w)
	for _, r := range v.m {
		if err := enc.Encode(r); err != nil {
			w.Close()
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return nil, err
	}
	if v.f, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *votes) get(key string, now time.Time) (*proto.Vote, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	r, ok := v.m[key]
	if !ok || r.Until < now.Unix() {
		return nil, false
	}
	vote := r.Vote
	return &vote, true
}

// put records vote under key unless one is already there, and returns the vote that
// stands (the earlier one wins, so two concurrent requests get the same answer).
func (v *votes) put(key string, until time.Time, vote proto.Vote, now time.Time) (proto.Vote, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if r, ok := v.m[key]; ok && r.Until >= now.Unix() {
		return r.Vote, nil
	}
	r := voteRec{Key: key, Until: until.Unix(), Vote: vote}
	if v.f != nil {
		b, err := json.Marshal(r)
		if err != nil {
			return proto.Vote{}, err
		}
		if _, err := v.f.Write(append(b, '\n')); err != nil {
			return proto.Vote{}, err
		}
		// The vote must survive a crash before anyone can see it.
		if err := v.f.Sync(); err != nil {
			return proto.Vote{}, err
		}
	}
	v.m[key] = r
	if len(v.m)%1024 == 0 {
		for k, x := range v.m {
			if x.Until < now.Unix() {
				delete(v.m, k)
			}
		}
	}
	return vote, nil
}
