package client

import (
	"sync"

	"github.com/libr-forum/Libr/core/libr/types"
)

// Events lets the UI layer (Wails on desktop, the native bridge on mobile) hear about
// background moderation work without the client depending on either.
type Events interface {
	// CronStatus reports the pending moderation queue after it changes.
	CronStatus(queue []map[string]interface{})
	// ModerationFinalized reports a pending item that reached a majority decision.
	ModerationFinalized(f types.FinalizedModeration)
}

var (
	eventsMu sync.RWMutex
	events   Events
)

// SetEvents registers the UI event sink; nil disables events.
func SetEvents(e Events) {
	eventsMu.Lock()
	events = e
	eventsMu.Unlock()
}

func emit(f func(Events)) {
	eventsMu.RLock()
	e := events
	eventsMu.RUnlock()
	if e != nil {
		f(e)
	}
}
