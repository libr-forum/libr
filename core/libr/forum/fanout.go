package forum

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

// fanout tries cands in order, `workers` at a time, until `want` attempts succeeded or
// the candidates run out. It returns the success count as soon as `enough` succeeded
// (or everything finished); the remaining attempts carry on and done closes after.
func fanout(ctx context.Context, cands []peer.ID, workers, want, enough int, try func(context.Context, peer.ID) bool) (int, <-chan struct{}) {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	next, ok, inflight := 0, 0, 0
	reached := make(chan struct{})
	var once sync.Once
	signal := func() { once.Do(func() { close(reached) }) }

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			for {
				// Wait while in-flight attempts could still fill the quota: if they
				// fail, more candidates are needed; if not, we are done.
				for ok+inflight >= want && inflight > 0 {
					cond.Wait()
				}
				if ok >= want || next >= len(cands) || ctx.Err() != nil {
					cond.Broadcast()
					return
				}
				p := cands[next]
				next++
				inflight++
				mu.Unlock()
				success := try(ctx, p)
				mu.Lock()
				inflight--
				if success {
					ok++
					if ok >= enough {
						signal()
					}
				}
				cond.Broadcast()
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		signal()
		close(done)
	}()
	<-reached
	mu.Lock()
	defer mu.Unlock()
	return ok, done
}

// parallel runs fn(0..n-1) with at most limit at once and waits for all.
func parallel(n, limit int, fn func(i int)) {
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer func() { <-sem; wg.Done() }()
			fn(i)
		}(i)
	}
	wg.Wait()
}
