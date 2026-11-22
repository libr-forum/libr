// Package forum is the LIBR v2 author/reader client: it collects a moderator quorum
// for a post, stores the certificate on the K closest DB nodes, indexes it in its
// community's time bucket, and reads feeds back, verifying every certificate.
package forum

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	kb "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/libr-forum/Libr/core/libr/proto"
	"github.com/libr-forum/Libr/core/libr/wire"
)

const (
	// Replicas is K: each certificate and index entry goes to the K closest DB nodes.
	Replicas = 4
	// StoredQuorum is ⌈(K+1)/2⌉ verified receipts, after which a post counts as stored.
	StoredQuorum = Replicas/2 + 1
	// indexReaders is how many DB nodes a feed reads each bucket from.
	indexReaders = 2
	// MaxFeedBuckets bounds one Feed call (one week of 15-minute buckets).
	MaxFeedBuckets = 7 * 24 * 4
	parallelGets   = 8
	lookupTimeout  = 30 * time.Second
	backgroundTime = 2 * wire.Timeout
)

var (
	// ErrRejected: enough moderators rejected the post that it cannot reach quorum.
	ErrRejected = errors.New("rejected by moderators")
	// ErrNoQuorum: too few moderators answered.
	ErrNoQuorum = errors.New("not enough moderators answered")
	// ErrNotStored: fewer than StoredQuorum DB nodes returned a valid receipt.
	ErrNotStored = errors.New("not enough storage receipts")
)

// Event is one network request, for measurements (librctl --trace writes JSON lines).
type Event struct {
	Op    string  `json:"op"` // vote | store | index_put | index_get | get
	Peer  string  `json:"peer"`
	Ms    float64 `json:"ms"`
	Bytes int     `json:"bytes"`
	Err   string  `json:"err,omitempty"`
}

// Client posts and reads as the author Key.
type Client struct {
	Host  host.Host
	DHT   *dht.IpfsDHT
	Sets  *proto.Registry
	Key   ed25519.PrivateKey
	Trace func(Event)      // optional; called concurrently
	Now   func() time.Time // nil means time.Now
}

// Posted is the outcome of Submit.
type Posted struct {
	Cert     proto.MsgCert
	Hash     proto.Hash
	Receipts []proto.StoreReceipt // cert receipts at return (>= StoredQuorum)
	Indexed  []proto.StoreReceipt // index receipts at return (>= StoredQuorum)
	// Done closes once replication to the remaining closest nodes has finished.
	Done <-chan struct{}
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Submit publishes content to community. It returns once a moderator quorum approved
// it and StoredQuorum nodes signed receipts for both the cert and its index entry.
func (c *Client) Submit(ctx context.Context, community string, content []byte) (*Posted, error) {
	set, ok := c.Sets.Head(community)
	if !ok {
		return nil, fmt.Errorf("unknown community %q", community)
	}
	if len(content) == 0 || len(content) > proto.MaxContent {
		return nil, fmt.Errorf("content must be 1..%d bytes", proto.MaxContent)
	}
	cert := proto.SignMsg(c.Key, proto.Msg{Community: community, Ts: c.now().Unix(), Content: content})
	cert.Epoch = set.Epoch
	unvoted := cert // late answers still read it after quorum returns; cert gets the votes
	votes, err := c.quorum(ctx, set, func(ctx context.Context, m proto.ModEntry) (*proto.Vote, error) {
		return c.askVote(ctx, m, unvoted)
	})
	if err != nil {
		return nil, err
	}
	cert.Votes = votes
	if err := proto.VerifyMsgCert(c.Sets, cert, c.now()); err != nil {
		return nil, fmt.Errorf("assembled certificate does not verify: %w", err)
	}

	p := &Posted{Cert: cert, Hash: cert.Hash()}
	var err2 error
	p.Done, err2 = both(
		func() (d <-chan struct{}, err error) {
			p.Receipts, d, err = c.replicate(ctx, "store", wire.ProtoStore, p.Hash, cert)
			return d, err
		},
		func() (d <-chan struct{}, err error) {
			p.Indexed, d, err = c.replicate(ctx, "index_put", wire.ProtoIndexPut, indexKeyOf(cert.Msg), cert)
			return d, err
		})
	return p, err2
}

func indexKeyOf(m proto.Msg) proto.Hash { return wire.IndexKey(m.Community, wire.Bucket(m.Ts)) }

// both runs a and b concurrently; done closes when both background parts have.
func both(a, b func() (<-chan struct{}, error)) (<-chan struct{}, error) {
	var wg sync.WaitGroup
	var dones [2]<-chan struct{}
	var errs [2]error
	for i, f := range []func() (<-chan struct{}, error){a, b} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dones[i], errs[i] = f()
		}()
	}
	wg.Wait()
	done := make(chan struct{})
	go func() {
		for _, d := range dones {
			if d != nil {
				<-d
			}
		}
		close(done)
	}()
	return done, errors.Join(errs[0], errs[1])
}

// quorum asks every moderator in parallel and returns as soon as Threshold approvals
// are in, or as soon as the outstanding answers can no longer reach it.
func (c *Client) quorum(ctx context.Context, set *proto.ModSet, ask func(context.Context, proto.ModEntry) (*proto.Vote, error)) ([]proto.Vote, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type answer struct {
		v   *proto.Vote
		err error
	}
	answers := make(chan answer, len(set.Mods))
	for _, m := range set.Mods {
		go func(m proto.ModEntry) {
			v, err := ask(ctx, m)
			answers <- answer{v, err}
		}(m)
	}
	var approvals []proto.Vote
	rejected, failed := 0, 0
	for range set.Mods {
		a := <-answers
		switch {
		case a.err != nil:
			failed++
		case !a.v.Approve:
			rejected++
		default:
			approvals = append(approvals, *a.v)
		}
		if uint64(len(approvals)) >= set.Threshold {
			return approvals, nil
		}
		if uint64(len(set.Mods)-rejected-failed) < set.Threshold {
			if uint64(len(set.Mods)-rejected) < set.Threshold {
				return nil, ErrRejected
			}
			return nil, fmt.Errorf("%w: %d approved, %d rejected, %d unreachable, need %d",
				ErrNoQuorum, len(approvals), rejected, failed, set.Threshold)
		}
	}
	return nil, ErrNoQuorum // unreachable: the loop decides by the last answer
}

func (c *Client) askVote(ctx context.Context, m proto.ModEntry, cert proto.MsgCert) (*proto.Vote, error) {
	p, err := peer.Decode(m.PeerID)
	if err != nil {
		return nil, err
	}
	var resp wire.VoteResp
	err = c.call(ctx, "vote", p, wire.ProtoVote, cert, &resp, func() error {
		if resp.Err != "" {
			return errors.New(resp.Err)
		}
		if resp.Vote == nil || !proto.VerifyVote(*resp.Vote, m.PubKey, cert.Msg, cert.Epoch) {
			return errors.New("invalid vote")
		}
		return nil
	})
	return resp.Vote, err
}

// spread runs try on the Replicas DB nodes closest to key, moving on to the
// next-closest node whenever one fails. It returns once StoredQuorum succeeded; the
// rest continue in the background until done closes.
func (c *Client) spread(ctx context.Context, op string, key proto.Hash, try func(context.Context, peer.ID) bool) (<-chan struct{}, error) {
	cands, err := c.closest(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), backgroundTime)
	ok, done := fanout(bg, cands, Replicas, Replicas, StoredQuorum, try)
	go func() { <-done; cancel() }()
	if ok < StoredQuorum {
		return done, fmt.Errorf("%s: %w (%d of %d)", op, ErrNotStored, ok, StoredQuorum)
	}
	return done, nil
}

// replicate sends cert on pid to the nodes closest to key and returns the verified
// receipts it had when the quorum was reached.
func (c *Client) replicate(ctx context.Context, op string, pid protocol.ID, key proto.Hash, cert proto.MsgCert) ([]proto.StoreReceipt, <-chan struct{}, error) {
	want := cert.Hash()
	var mu sync.Mutex
	var receipts []proto.StoreReceipt
	done, err := c.spread(ctx, op, key, func(ctx context.Context, p peer.ID) bool {
		var resp wire.ReceiptResp
		err := c.call(ctx, op, p, pid, cert, &resp, func() error {
			if resp.Err != "" {
				return errors.New(resp.Err)
			}
			return checkReceipt(resp.Receipt, want, p)
		})
		if err == nil {
			mu.Lock()
			receipts = append(receipts, *resp.Receipt)
			mu.Unlock()
		}
		return err == nil
	})
	mu.Lock()
	defer mu.Unlock()
	return append([]proto.StoreReceipt(nil), receipts...), done, err
}

// checkReceipt accepts only a receipt for this cert signed by the key of the peer we
// actually talked to, so a node cannot hand in receipts on another node's behalf.
func checkReceipt(r *proto.StoreReceipt, want proto.Hash, from peer.ID) error {
	if r == nil || !r.Verify() || r.CertHash != want {
		return errors.New("invalid receipt")
	}
	if id, err := proto.PeerIDFor(r.Node); err != nil || id != from.String() {
		return errors.New("receipt signed by another node")
	}
	return nil
}

// Feed returns the verified posts of community with timestamps in [since, until],
// newest first, one per message ID. Buckets that no DB node answered for are skipped;
// the error is set only if none could be read.
func (c *Client) Feed(ctx context.Context, community string, since, until time.Time) ([]proto.MsgCert, error) {
	if _, ok := c.Sets.Head(community); !ok {
		return nil, fmt.Errorf("unknown community %q", community)
	}
	first, last := wire.Bucket(since.Unix()), wire.Bucket(until.Unix())
	if n := (last-first)/wire.BucketSeconds + 1; n > MaxFeedBuckets || n < 1 {
		return nil, fmt.Errorf("feed range of %d buckets out of range 1..%d", n, MaxFeedBuckets)
	}

	var mu sync.Mutex
	hashes := map[proto.Hash]bool{}
	tombs := map[proto.Hash][]proto.Tombstone{}
	read := 0
	var lastErr error
	var buckets []int64
	for b := first; b <= last; b += wire.BucketSeconds {
		buckets = append(buckets, b)
	}
	parallel(len(buckets), parallelGets, func(i int) {
		hs, ts, err := c.indexGet(ctx, community, buckets[i])
		mu.Lock()
		defer mu.Unlock()
		if err == nil {
			read++
		} else {
			lastErr = err
		}
		for _, h := range hs {
			hashes[h] = true
		}
		for _, t := range ts {
			tombs[t.MsgID] = append(tombs[t.MsgID], t)
		}
	})
	if read == 0 {
		return nil, fmt.Errorf("no DB node answered for the feed index: %w", lastErr)
	}

	list := make([]proto.Hash, 0, len(hashes))
	for h := range hashes {
		list = append(list, h)
	}
	byID := map[proto.Hash]proto.MsgCert{}
	parallel(len(list), parallelGets, func(i int) {
		cert, err := c.Get(ctx, list[i])
		if err != nil || cert.Msg.Community != community || cert.Msg.Ts < since.Unix() || cert.Msg.Ts > until.Unix() {
			return
		}
		for _, t := range tombs[cert.Msg.ID()] { // unverified until checked against the message
			if proto.VerifyTombstone(c.Sets, t, cert.Msg) == nil {
				return
			}
		}
		mu.Lock()
		byID[cert.Msg.ID()] = cert
		mu.Unlock()
	})
	out := make([]proto.MsgCert, 0, len(byID))
	for _, cert := range byID {
		out = append(out, cert)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Msg.Ts != out[j].Msg.Ts {
			return out[i].Msg.Ts > out[j].Msg.Ts
		}
		return string(out[i].AuthorSig) < string(out[j].AuthorSig)
	})
	return out, nil
}

// indexGet unions one bucket's hashes and tombstones from indexReaders of the
// closest nodes.
func (c *Client) indexGet(ctx context.Context, community string, bucket int64) ([]proto.Hash, []proto.Tombstone, error) {
	cands, err := c.closest(ctx, wire.IndexKey(community, bucket))
	if err != nil {
		return nil, nil, fmt.Errorf("lookup: %w", err)
	}
	var mu sync.Mutex
	seen := map[proto.Hash]bool{}
	var tombs []proto.Tombstone
	req := wire.IndexGetReq{Community: community, Bucket: bucket}
	ok, _ := fanout(ctx, cands, indexReaders, indexReaders, indexReaders, func(ctx context.Context, p peer.ID) bool {
		var resp wire.IndexGetResp
		err := c.call(ctx, "index_get", p, wire.ProtoIndexGet, req, &resp, func() error {
			if resp.Err != "" {
				return errors.New(resp.Err)
			}
			if len(resp.Hashes) > wire.MaxIndexHashes || len(resp.Tombs) > wire.MaxIndexHashes {
				return errors.New("oversized index reply")
			}
			return nil
		})
		if err == nil {
			mu.Lock()
			for _, h := range resp.Hashes {
				seen[h] = true
			}
			tombs = append(tombs, resp.Tombs...)
			mu.Unlock()
		}
		return err == nil
	})
	mu.Lock()
	defer mu.Unlock()
	out := make([]proto.Hash, 0, len(seen))
	for h := range seen {
		out = append(out, h)
	}
	if ok == 0 {
		return nil, nil, fmt.Errorf("none of %d closest nodes answered", len(cands))
	}
	return out, tombs, nil
}

// Get fetches the cert stored under h from the closest nodes, accepting the first
// copy that hashes to h and verifies; tampered or forged copies are skipped. It does
// not check for removal (Feed does, through the bucket's tombstones).
func (c *Client) Get(ctx context.Context, h proto.Hash) (proto.MsgCert, error) {
	cands, err := c.closest(ctx, h)
	if err != nil {
		return proto.MsgCert{}, err
	}
	var cert proto.MsgCert
	ok, _ := fanout(ctx, cands, 1, 1, 1, func(ctx context.Context, p peer.ID) bool {
		var resp wire.GetResp
		return c.call(ctx, "get", p, wire.ProtoGet, wire.GetReq{Hash: h}, &resp, func() error {
			switch {
			case resp.Err != "":
				return errors.New(resp.Err)
			case resp.Cert == nil || resp.Cert.Hash() != h:
				return errors.New("cert does not match its hash")
			}
			if err := proto.VerifyMsgCert(c.Sets, *resp.Cert, c.now()); err != nil {
				return fmt.Errorf("cert does not verify: %w", err)
			}
			cert = *resp.Cert
			return nil
		}) == nil
	})
	if ok != 1 {
		return proto.MsgCert{}, errors.New("no DB node returned a valid copy")
	}
	return cert, nil
}

// closest returns the DB nodes closest to key, nearest first.
func (c *Client) closest(ctx context.Context, key proto.Hash) ([]peer.ID, error) {
	ctx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()
	ps, err := c.DHT.GetClosestPeers(ctx, string(key[:]))
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, errors.New("no DB nodes found")
	}
	return kb.SortClosestPeers(ps, kb.ConvertKey(string(key[:]))), nil
}

// call connects to p (looking it up on the DHT if needed), makes one request, runs
// check on the reply and traces the outcome.
func (c *Client) call(ctx context.Context, op string, p peer.ID, pid protocol.ID, req, resp any, check func() error) error {
	start := time.Now()
	n := 0
	err := c.connect(ctx, p)
	if err == nil {
		n, err = wire.CallN(ctx, c.Host, p, pid, req, resp)
	}
	if err == nil {
		err = check()
	}
	if c.Trace != nil {
		e := Event{Op: op, Peer: p.String(), Ms: float64(time.Since(start).Microseconds()) / 1000, Bytes: n}
		if err != nil {
			e.Err = err.Error()
		}
		c.Trace(e)
	}
	return err
}

func (c *Client) connect(ctx context.Context, p peer.ID) error {
	if c.Host.Network().Connectedness(p) != network.NotConnected {
		return nil
	}
	if len(c.Host.Peerstore().Addrs(p)) > 0 {
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := c.Host.Connect(cctx, peer.AddrInfo{ID: p})
		cancel()
		if err == nil {
			return nil
		}
	}
	ctx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()
	ai, err := c.DHT.FindPeer(ctx, p)
	if err != nil {
		return fmt.Errorf("find %s: %w", p.ShortString(), err)
	}
	return c.Host.Connect(ctx, ai)
}
