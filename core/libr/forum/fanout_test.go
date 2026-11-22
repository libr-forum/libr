package forum

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
)

func ids(n int) []peer.ID {
	out := make([]peer.ID, n)
	for i := range out {
		out[i] = peer.ID(fmt.Sprint(i))
	}
	return out
}

func TestFanout(t *testing.T) {
	failing := func(bad ...int) func(context.Context, peer.ID) bool {
		return func(_ context.Context, p peer.ID) bool {
			for _, b := range bad {
				if p == peer.ID(fmt.Sprint(b)) {
					return false
				}
			}
			return true
		}
	}
	for _, tc := range []struct {
		name                 string
		cands, workers, want int
		bad                  []int
		wantOK               int
	}{
		{"all succeed: stops at want", 8, 4, 4, nil, 4},
		{"failures move to next candidates", 8, 4, 4, []int{0, 2}, 4},
		{"not enough candidates", 5, 4, 4, []int{0, 1}, 3},
		{"everything fails", 3, 2, 2, []int{0, 1, 2}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var tried atomic.Int32
			try := failing(tc.bad...)
			ok, done := fanout(context.Background(), ids(tc.cands), tc.workers, tc.want, tc.want, func(ctx context.Context, p peer.ID) bool {
				tried.Add(1)
				return try(ctx, p)
			})
			<-done
			if ok != tc.wantOK {
				t.Fatalf("ok = %d, want %d", ok, tc.wantOK)
			}
			if max := tc.want + len(tc.bad); int(tried.Load()) > max && int(tried.Load()) > tc.workers {
				t.Errorf("tried %d candidates, expected at most %d", tried.Load(), max)
			}
		})
	}
}

// TestFanoutReturnsAtEnough: the caller gets control back at `enough` while a slow
// attempt is still running.
func TestFanoutReturnsAtEnough(t *testing.T) {
	release := make(chan struct{})
	ok, done := fanout(context.Background(), ids(4), 4, 4, 3, func(_ context.Context, p peer.ID) bool {
		if p == "3" {
			<-release
		}
		return true
	})
	if ok != 3 {
		t.Fatalf("ok at return = %d, want 3", ok)
	}
	select {
	case <-done:
		t.Fatal("done closed before the slow attempt finished")
	default:
	}
	close(release)
	<-done
}
