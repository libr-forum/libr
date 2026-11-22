// Package wire is the LIBR v2 request/response layer: one request and one response per
// libp2p stream, each a uvarint length followed by JSON, with size and rate limits.
package wire

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"golang.org/x/time/rate"
)

// MaxFrame bounds one message: a 1 MiB post plus certificate overhead.
const MaxFrame = 2 << 20

// Timeout bounds one whole request/response exchange.
const Timeout = 30 * time.Second

// ErrTooLarge is returned for frames over MaxFrame.
var ErrTooLarge = errors.New("wire: frame too large")

// WriteFrame writes v as a length-prefixed JSON frame.
func WriteFrame(w io.Writer, v any) error {
	_, err := writeFrame(w, v)
	return err
}

func writeFrame(w io.Writer, v any) (int, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	if len(b) > MaxFrame {
		return 0, ErrTooLarge
	}
	buf := binary.AppendUvarint(make([]byte, 0, len(b)+binary.MaxVarintLen64), uint64(len(b)))
	return w.Write(append(buf, b...))
}

// ReadFrame reads one frame into v, refusing frames over MaxFrame before reading them.
func ReadFrame(r *bufio.Reader, v any) error {
	_, err := readFrame(r, v)
	return err
}

func readFrame(r *bufio.Reader, v any) (int, error) {
	n, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}
	if n > MaxFrame {
		return 0, ErrTooLarge
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return 0, err
	}
	return len(b) + uvarintLen(n), json.Unmarshal(b, v)
}

func uvarintLen(n uint64) int { return len(binary.AppendUvarint(nil, n)) }

// Call sends req to p on proto and decodes the reply into resp. Relayed connections are
// allowed: NATed peers are only reachable through circuits.
func Call(ctx context.Context, h host.Host, p peer.ID, proto protocol.ID, req, resp any) error {
	_, err := CallN(ctx, h, p, proto, req, resp)
	return err
}

// CallN is Call that also reports the bytes sent plus received (for measurements).
func CallN(ctx context.Context, h host.Host, p peer.ID, proto protocol.ID, req, resp any) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	s, err := h.NewStream(network.WithAllowLimitedConn(ctx, "libr"), p, proto)
	if err != nil {
		return 0, err
	}
	defer s.Close()
	if d, ok := ctx.Deadline(); ok {
		_ = s.SetDeadline(d)
	}
	sent, err := writeFrame(s, req)
	if err != nil {
		_ = s.Reset()
		return sent, err
	}
	if err := s.CloseWrite(); err != nil {
		return sent, err
	}
	got, err := readFrame(bufio.NewReader(s), resp)
	return sent + got, err
}

// Handle serves proto: each request is decoded into Req, passed with the authenticated
// remote peer to fn, and fn's reply is written back. Requests over lim's per-peer rate
// get reject() instead; bodies are never logged.
func Handle[Req, Resp any](h host.Host, proto protocol.ID, lim *Limiter, fn func(from peer.ID, req Req) Resp, reject func(error) Resp) {
	h.SetStreamHandler(proto, func(s network.Stream) {
		defer s.Close()
		_ = s.SetDeadline(time.Now().Add(Timeout))
		from := s.Conn().RemotePeer()
		var resp Resp
		var req Req
		if !lim.Allow(string(from)) { // before reading, so a flood costs us no body reads
			resp = reject(ErrRateLimited)
		} else if err := ReadFrame(bufio.NewReader(s), &req); err != nil {
			resp = reject(fmt.Errorf("bad request: %w", err))
		} else {
			resp = fn(from, req)
		}
		if err := WriteFrame(s, resp); err != nil {
			_ = s.Reset()
		}
	})
}

// ErrRateLimited is the rejection for peers over their request rate.
var ErrRateLimited = errors.New("ratelimited")

// Limiter is a token bucket per key (a peer ID, or an author key).
type Limiter struct {
	mu    sync.Mutex
	every rate.Limit
	burst int
	peers map[string]*rate.Limiter
}

// maxTracked bounds memory against peers that rotate identities.
const maxTracked = 10_000

// NewLimiter allows perSecond requests per peer with the given burst.
func NewLimiter(perSecond float64, burst int) *Limiter {
	return &Limiter{every: rate.Limit(perSecond), burst: burst, peers: map[string]*rate.Limiter{}}
}

// Allow reports whether key may make another request now.
func (l *Limiter) Allow(p string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.peers[p]
	if !ok {
		if len(l.peers) >= maxTracked {
			// shortcut: forget everyone rather than track LRU; fine until sustained >10k peers
			l.peers = map[string]*rate.Limiter{}
		}
		lim = rate.NewLimiter(l.every, l.burst)
		l.peers[p] = lim
	}
	return lim.Allow()
}
