package core

import (
	"context"
	"crypto/ed25519"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	cache "github.com/libr-forum/Libr/core/mod_client/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/identity"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------------------------------------------------------------------------
// Moderation retry cron
// ---------------------------------------------------------------------------

var (
	modCronMu     sync.Mutex
	modCronCancel context.CancelFunc
	WailsCtx      context.Context
)

// MaybeStartCron starts the moderation retry cron if there are pending files.
func MaybeStartCron() {
	pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, _ := filepath.Glob(pattern)
	if len(files) > 0 {
		StartModerationCron()
	}
}

// StartModerationCron starts a background goroutine that retries pending
// moderations every 60 s. It is idempotent: calling it while already running
// is a no-op. The goroutine stops itself once there are no pending files left.
func StartModerationCron() {
	modCronMu.Lock()
	if modCronCancel != nil {
		modCronMu.Unlock()
		return // already running
	}
	ctx, cancel := context.WithCancel(context.Background())
	modCronCancel = cancel
	modCronMu.Unlock()

	log.Println("🚀 Starting moderation retry cron...")

	go func() {
		defer func() {
			modCronMu.Lock()
			modCronCancel = nil
			modCronMu.Unlock()
			log.Println("🛑 Moderation retry cron stopped")
		}()

		// Process immediately on start, then every 60 s.
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		RetryPendingModerations()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
				files, err := filepath.Glob(pattern)
				if err != nil {
					log.Printf("Cron check error: %v", err)
					continue
				}
				if len(files) == 0 {
					log.Println("✅ All moderations resolved — stopping cron")
					return
				}
				RetryPendingModerations()
			}
		}
	}()
}

// StopModerationCron cancels the moderation retry cron if running.
func StopModerationCron() {
	modCronMu.Lock()
	defer modCronMu.Unlock()
	if modCronCancel != nil {
		modCronCancel()
		// modCronCancel is cleared by the goroutine's defer
	}
}

// IsModerationCronRunning reports whether the moderation retry cron is active.
func IsModerationCronRunning() bool {
	modCronMu.Lock()
	defer modCronMu.Unlock()
	return modCronCancel != nil
}

// RetryPendingModerations reads all pending moderation files, retries sending
// to mods that are now online, tallies votes and finalises or re-saves each entry.
func RetryPendingModerations() {
	pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.LogToFile("[DEBUG] Failed to list pending moderation files")
		log.Printf("Failed to list pending moderation files: %v", err)
		return
	}
	if len(files) == 0 {
		return
	}

	var pendingQueue []map[string]interface{}

	// Fetch current online mods once; index by public key for O(1) lookup.
	allMods, err := util.GetOnlineMods()
	if err != nil {
		log.Printf("⚠️ GetOnlineMods failed: %v — will retry next tick", err)
		return
	}
	modByPubKey := make(map[string]types.Mod, len(allMods))
	for _, m := range allMods {
		modByPubKey[m.PublicKey] = m
	}

	for _, filePath := range files {
		pending, err := cache.LoadPendingModeration(filePath)
		if err != nil {
			logger.LogToFile("[DEBUG] Could not load pending file")
			log.Printf("Could not load pending file %s: %v", filePath, err)
			continue
		}

		// Resolve AwaitingMods to their current network addresses.
		var retryMods []types.Mod
		for _, pubKey := range pending.AwaitingMods {
			if mod, ok := modByPubKey[pubKey]; ok {
				retryMods = append(retryMods, mod)
			}
		}

		var newCerts []types.ModCert
		if len(retryMods) > 0 {
			newCerts = ManualSendToMods(pending.MsgCert, retryMods, "", false)
		}

		// Merge new certs with what we already have.
		allCerts := append(pending.PartialCerts, newCerts...)

		// Track which awaiting mods gave a final decision this round.
		respondedSet := make(map[string]struct{}, len(newCerts))
		for _, mc := range newCerts {
			if mc.Status != "acknowledged" {
				respondedSet[mc.PublicKey] = struct{}{}
			}
		}
		var newAwaiting []string
		for _, pub := range pending.AwaitingMods {
			if _, responded := respondedSet[pub]; !responded {
				newAwaiting = append(newAwaiting, pub)
			}
		}

		// Tally votes across all collected certs.
		var rejCount, accCount int
		for _, cert := range allCerts {
			switch cert.Status {
			case "0":
				rejCount++
			case "1":
				accCount++
			}
		}
		totalDecisions := rejCount + accCount
		totalMods := len(allCerts)

		log.Printf("[Cron] %s — total=%d acc=%d rej=%d awaiting=%d ackCount=%d",
			pending.MsgSign, totalMods, accCount, rejCount, len(newAwaiting), pending.AckCount)

		switch {
		case totalDecisions > 0 && rejCount > totalDecisions/2:
			// Majority rejected.
			log.Printf("❌ Majority rejected — deleting %s", pending.MsgSign)
			if WailsCtx != nil {
				runtime.EventsEmit(WailsCtx, "moderation_finalized", map[string]interface{}{"status": "rejected", "id": pending.MsgSign})
			}
			cache.DeletePendingModeration(pending.MsgSign)
			maybeCleanupIdentity(pending.SignerIdentityID)

		case totalDecisions > 0 && accCount > totalDecisions/2:
			// Majority approved.
			log.Printf("✅ Majority approved — processing %s", pending.MsgSign)
			if WailsCtx != nil {
				runtime.EventsEmit(WailsCtx, "moderation_finalized", map[string]interface{}{"status": "approved", "id": pending.MsgSign})
			}
			tsMin := pending.MsgCert.Msg.Ts - (pending.MsgCert.Msg.Ts % 60)
			key := util.GenerateNodeID(strconv.FormatInt(tsMin, 10))

			if pending.MsgCert.Type == "manual_mod" || pending.MsgCert.Reason == "Image attached" {
				// New message with image: forward to chat DB normally.
				privKey, err := identity.LoadPrivateKey(pending.SignerIdentityID)
				if err != nil {
					log.Printf("Failed to load signer identity %s for %s: %v", pending.SignerIdentityID, pending.MsgSign, err)
					continue
				}
				msgCert := CreateMsgCertWithPrivateKey(pending.MsgCert.Msg.Content, pending.MsgCert.Msg.Ts, allCerts, privKey)
				SendToDb(key, msgCert, "/route=store")
			} else {
				// Report to delete an existing message.
				repCert := CreateRepCert(pending.MsgCert, allCerts, "report")
				SendToDb(key, repCert, "/route=delete")
			}
			cache.DeletePendingModeration(pending.MsgSign)
			maybeCleanupIdentity(pending.SignerIdentityID)

		default:
			// No clear majority yet, or pending more votes.
			log.Printf("⏳ Awaiting more responses for %s", pending.MsgSign)
			pending.PartialCerts = allCerts
			pending.AwaitingMods = newAwaiting
			_ = cache.SavePendingModeration(pending)

			// Add to frontend queue display
			pendingQueue = append(pendingQueue, map[string]interface{}{
				"id":           pending.MsgSign,
				"ts":           pending.MsgCert.Msg.Ts,
				"content":      pending.MsgCert.Msg.Content,
				"reason":       pending.MsgCert.Reason,
				"totalMods":    pending.AckCount,
				"ackCount":     accCount + rejCount,
				"approved":     accCount,
				"rejected":     rejCount,
				"awaitingMods": len(newAwaiting),
			})
		}
	}

	if WailsCtx != nil {
		runtime.EventsEmit(WailsCtx, "cron_status_update", pendingQueue)
	}
}

func maybeCleanupIdentity(identityID string) {
	if identityID == "" || identityID == identity.MainIdentityID {
		return
	}
	if cache.HasPendingModerationForIdentity(identityID) {
		return
	}
	activeID, err := identity.GetActiveIdentityID()
	if err == nil && activeID == identityID {
		return
	}
	if err := identity.DeleteIdentity(identityID); err != nil {
		log.Printf("Failed to cleanup identity %s: %v", identityID, err)
	}
}

func buildPendingQueueSnapshot() ([]string, []map[string]interface{}, error) {
	pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, err
	}

	var pendingQueue []map[string]interface{}
	for _, filePath := range files {
		pending, err := cache.LoadPendingModeration(filePath)
		if err != nil {
			logger.LogToFile("[DEBUG] Could not load pending file")
			log.Printf("Could not load pending file %s: %v", filePath, err)
			continue
		}

		var approved, rejected int
		for _, cert := range pending.PartialCerts {
			switch cert.Status {
			case "1":
				approved++
			case "0":
				rejected++
			}
		}

		pendingQueue = append(pendingQueue, map[string]interface{}{
			"id":           pending.MsgSign,
			"ts":           pending.MsgCert.Msg.Ts,
			"content":      pending.MsgCert.Msg.Content,
			"reason":       pending.MsgCert.Reason,
			"totalMods":    pending.AckCount,
			"ackCount":     approved + rejected,
			"approved":     approved,
			"rejected":     rejected,
			"awaitingMods": len(pending.AwaitingMods),
		})
	}

	return files, pendingQueue, nil
}

// ---------------------------------------------------------------------------
// Mod presence refresh cron
// ---------------------------------------------------------------------------

var (
	refreshCronCancel context.CancelFunc
	refreshCronMu     sync.Mutex
)

// StartRefreshCron begins a background goroutine that calls RefreshModPresence
// every intervalSec seconds, keeping this mod's discovery entry alive.
// Calling it while already running is a no-op.
func StartRefreshCron(pubKeyB64 string, privKey ed25519.PrivateKey, intervalSec int) {
	refreshCronMu.Lock()
	defer refreshCronMu.Unlock()

	if refreshCronCancel != nil {
		return // already running
	}

	ctx, cancel := context.WithCancel(context.Background())
	refreshCronCancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := util.RefreshModPresence(pubKeyB64, privKey); err != nil {
					log.Printf("⚠️ RefreshModPresence failed: %v", err)
				} else {
					log.Println("🔄 Mod presence refreshed")
				}
			case <-ctx.Done():
				log.Println("🔴 Mod presence refresh cron stopped")
				return
			}
		}
	}()

	log.Printf("⏱  Mod presence refresh cron started (interval: %ds)", intervalSec)
}

// StopRefreshCron stops the mod presence refresh goroutine.
func StopRefreshCron() {
	refreshCronMu.Lock()
	defer refreshCronMu.Unlock()

	if refreshCronCancel != nil {
		refreshCronCancel()
		refreshCronCancel = nil
	}
}
