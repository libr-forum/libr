package cache

import (
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/mod_client/types"
)

type sentMsg struct {
	cert      types.MsgCert
	expiresAt time.Time
}

var (
	sentCache   []sentMsg
	sentCacheMu sync.Mutex
)

const sentCacheTTL = 60 * time.Second

// AddToSentCache adds a message that was just sent to the in-memory cache.
// These messages are merged into FetchRecent results to prevent UI staleness.
func AddToSentCache(cert types.MsgCert) {
	sentCacheMu.Lock()
	defer sentCacheMu.Unlock()

	sentCache = append(sentCache, sentMsg{
		cert:      cert,
		expiresAt: time.Now().Add(sentCacheTTL),
	})
}

// GetRecentSentMessages returns non-expired messages from the sent cache.
func GetRecentSentMessages() []types.MsgCert {
	sentCacheMu.Lock()
	defer sentCacheMu.Unlock()

	var valid []types.MsgCert
	var stillFresh []sentMsg
	now := time.Now()

	for _, m := range sentCache {
		if now.Before(m.expiresAt) {
			valid = append(valid, m.cert)
			stillFresh = append(stillFresh, m)
		}
	}
	sentCache = stillFresh
	return valid
}
