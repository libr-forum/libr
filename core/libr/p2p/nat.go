package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
)

// How often a private host checks its relays, and how long before a reservation
// expires it is renewed. A dead relay's circuit address is withdrawn within one tick.
const (
	keepTick      = 15 * time.Second
	renewMargin   = 5 * time.Minute
	reserveTimout = 15 * time.Second
)

// reservations keeps a private host's reservations on its relays and advertises one
// circuit address per live reservation and nothing else.
type reservations struct {
	h      host.Host
	relays []peer.AddrInfo

	mu     sync.Mutex
	expiry map[peer.ID]time.Time
}

func (r *reservations) addrs([]ma.Multiaddr) []ma.Multiaddr {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []ma.Multiaddr
	for _, ri := range r.relays {
		if exp, ok := r.expiry[ri.ID]; !ok || time.Now().After(exp) {
			continue
		}
		circuit := ma.StringCast("/p2p/" + ri.ID.String() + "/p2p-circuit")
		for _, a := range ri.Addrs {
			out = append(out, a.Encapsulate(circuit))
		}
	}
	return out
}

// refresh (re)reserves on every relay that is disconnected or close to expiry.
func (r *reservations) refresh(ctx context.Context) {
	var wg sync.WaitGroup
	for _, ri := range r.relays {
		r.mu.Lock()
		exp, ok := r.expiry[ri.ID]
		r.mu.Unlock()
		if ok && time.Until(exp) > renewMargin && r.h.Network().Connectedness(ri.ID) == network.Connected {
			continue
		}
		wg.Add(1)
		go func(ri peer.AddrInfo) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, reserveTimout)
			defer cancel()
			var res *client.Reservation
			err := r.h.Connect(ctx, ri)
			if err == nil {
				res, err = client.Reserve(ctx, r.h, ri)
			}
			r.mu.Lock()
			defer r.mu.Unlock()
			if err != nil {
				delete(r.expiry, ri.ID)
				return
			}
			r.h.ConnManager().Protect(ri.ID, "relay")
			r.expiry[ri.ID] = res.Expiration
		}(ri)
	}
	wg.Wait()
}

func (r *reservations) run(ctx context.Context) {
	t := time.NewTicker(keepTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.refresh(ctx)
		}
	}
}

// privateHost stops the reservation loop on Close.
type privateHost struct {
	host.Host
	cancel context.CancelFunc
}

func (h *privateHost) Close() error {
	h.cancel()
	return h.Host.Close()
}

// KeepFindable keeps a DHT-client host (a moderator) connected to the DB nodes closest
// to its own ID. Those servers answer FindPeer for it with its circuit addresses, so
// authors can reach it although it is not in any routing table. Runs until ctx ends.
func KeepFindable(ctx context.Context, h host.Host, d *dht.IpfsDHT, every time.Duration) {
	for {
		_ = Announce(ctx, h, d)
		select {
		case <-ctx.Done():
			return
		case <-time.After(every):
		}
	}
}

// Announce connects h to the DB nodes closest to its own peer ID and protects those
// connections.
func Announce(ctx context.Context, h host.Host, d *dht.IpfsDHT) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	closest, err := d.GetClosestPeers(ctx, string(h.ID()))
	if err != nil {
		return err
	}
	for _, p := range closest {
		if err := h.Connect(ctx, peer.AddrInfo{ID: p}); err == nil {
			h.ConnManager().Protect(p, "libr-announce")
		}
	}
	return nil
}

// Join connects to the bootstrap peers and returns once the routing table has been
// refreshed through at least one of them. Connect returns before the DHT has added the
// peer (that happens after identify), so refreshing straight away walks an empty table.
func Join(ctx context.Context, h host.Host, d *dht.IpfsDHT, boot []peer.AddrInfo) error {
	var errs []error
	for _, b := range boot {
		if err := h.Connect(ctx, b); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", b.ID.ShortString(), err))
		}
	}
	if len(errs) == len(boot) {
		return fmt.Errorf("no bootstrap peer reachable: %w", errors.Join(errs...))
	}
	for d.RoutingTable().Size() == 0 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("bootstrap peers never entered the routing table: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
	return <-d.RefreshRoutingTable()
}
