package core

import (
	"context"
	"crypto/ed25519"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	cache "github.com/libr-forum/Libr/core/mod_client/core/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/identity"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	peer "github.com/libr-forum/Libr/core/mod_client/peers"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

var (
	modCronMu     sync.Mutex
	modCronCancel context.CancelFunc

	notifHandler   NotificationHandler
	notifHandlerMu sync.Mutex
)

type NotificationHandler interface {
	OnNotification(title string, message string)
}

func SetNotificationHandler(h NotificationHandler) {
	notifHandlerMu.Lock()
	defer notifHandlerMu.Unlock()
	notifHandler = h
}

func TriggerNotification(title, message string) {
	notifHandlerMu.Lock()
	h := notifHandler
	notifHandlerMu.Unlock()
	if h != nil {
		h.OnNotification(title, message)
	}
}

func MaybeStartCron() {
	pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, _ := filepath.Glob(pattern)
	if len(files) > 0 {
		StartModerationCron()
	}
}

func StartModerationCron() {
	modCronMu.Lock()
	if modCronCancel != nil {
		modCronMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	modCronCancel = cancel
	modCronMu.Unlock()

	log.Println("Starting moderation retry cron (Mobile)...")

	go func() {
		defer func() {
			modCronMu.Lock()
			modCronCancel = nil
			modCronMu.Unlock()
			log.Println("Moderation retry cron stopped")
		}()

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
					log.Println("All moderations resolved - stopping cron")
					return
				}
				RetryPendingModerations()
			}
		}
	}()
}

func StopModerationCron() {
	modCronMu.Lock()
	defer modCronMu.Unlock()
	if modCronCancel != nil {
		modCronCancel()
	}
}

func IsModerationCronRunning() bool {
	modCronMu.Lock()
	defer modCronMu.Unlock()
	return modCronCancel != nil
}

func RetryPendingModerations() {
	if peer.Peer == nil || peer.Peer.Host == nil {
		log.Println("[Cron] Libp2p node not initialized yet, skipping moderation retry tick")
		return
	}

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

	allMods, err := util.GetOnlineMods()
	if err != nil {
		log.Printf("GetOnlineMods failed: %v - will retry next tick", err)
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

		allCerts := append(pending.PartialCerts, newCerts...)

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

		log.Printf("[Cron] %s - total=%d acc=%d rej=%d awaiting=%d ackCount=%d",
			pending.MsgSign, totalMods, accCount, rejCount, len(newAwaiting), pending.AckCount)

		switch {
		case totalDecisions > 0 && rejCount > totalDecisions/2:
			log.Printf("Majority rejected - finalizing %s", pending.MsgSign)
			itemType := "report"
			if pending.MsgCert.Type == "manual_mod" || pending.MsgCert.Reason == "Image attached" {
				itemType = "manual_mod"
			}
			tot := pending.AckCount
			if tot < totalDecisions {
				tot = totalDecisions
			}
			_ = cache.SaveFinalizedModeration(types.FinalizedModeration{
				MsgSign:     pending.MsgSign,
				Type:        itemType,
				Status:      "rejected",
				Approved:    accCount,
				Rejected:    rejCount,
				Total:       tot,
				FinalizedAt: time.Now(),
			})
			cache.DeletePendingModeration(pending.MsgSign)
			maybeCleanupIdentity(pending.SignerIdentityID)

			if itemType == "manual_mod" {
				TriggerNotification("Image Rejected", "Your image was rejected by moderators.")
			} else {
				TriggerNotification("Report Dismissed", "A reported message was reviewed and dismissed.")
			}

		case totalDecisions > 0 && accCount > totalDecisions/2:
			log.Printf("Majority approved - processing %s", pending.MsgSign)
			tsMin := pending.MsgCert.Msg.Ts - (pending.MsgCert.Msg.Ts % 60)
			key := util.GenerateNodeID(strconv.FormatInt(tsMin, 10))

			itemType := "report"
			if pending.MsgCert.Type == "manual_mod" || pending.MsgCert.Reason == "Image attached" {
				itemType = "manual_mod"
				privKey, err := identity.LoadPrivateKey(pending.SignerIdentityID)
				if err != nil {
					log.Printf("Failed to load signer identity %s for %s: %v", pending.SignerIdentityID, pending.MsgSign, err)
					continue
				}
				msgCert := CreateMsgCertWithPrivateKey(pending.MsgCert.Msg.Content, pending.MsgCert.Msg.Ts, allCerts, privKey)
				SendToDb(key, msgCert, "/route=store")
			} else {
				repCert := CreateRepCert(pending.MsgCert, allCerts, "report")
				SendToDb(key, repCert, "/route=delete")
			}
			tot := pending.AckCount
			if tot < totalDecisions {
				tot = totalDecisions
			}
			_ = cache.SaveFinalizedModeration(types.FinalizedModeration{
				MsgSign:     pending.MsgSign,
				Type:        itemType,
				Status:      "approved",
				Approved:    accCount,
				Rejected:    rejCount,
				Total:       tot,
				FinalizedAt: time.Now(),
			})
			cache.DeletePendingModeration(pending.MsgSign)
			maybeCleanupIdentity(pending.SignerIdentityID)

			if itemType == "manual_mod" {
				TriggerNotification("Image Approved", "Your image was approved by moderators and published to the feed.")
			} else {
				TriggerNotification("Report Upheld", "A reported message was verified as a violation and removed.")
			}

		default:
			log.Printf("Awaiting more responses for %s", pending.MsgSign)
			pending.PartialCerts = allCerts
			pending.AwaitingMods = newAwaiting
			_ = cache.SavePendingModeration(pending)
		}
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

var (
	refreshCronCancel context.CancelFunc
	refreshCronMu     sync.Mutex
)

func StartRefreshCron(pubKeyB64 string, privKey ed25519.PrivateKey, intervalSec int) {
	refreshCronMu.Lock()
	defer refreshCronMu.Unlock()

	if refreshCronCancel != nil {
		return
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
					log.Printf("RefreshModPresence failed: %v", err)
				} else {
					log.Println("Mod presence refreshed")
				}
			case <-ctx.Done():
				log.Println("Mod presence refresh cron stopped")
				return
			}
		}
	}()

	log.Printf("Mod presence refresh cron started (interval: %ds)", intervalSec)
}

func StopRefreshCron() {
	refreshCronMu.Lock()
	defer refreshCronMu.Unlock()

	if refreshCronCancel != nil {
		refreshCronCancel()
		refreshCronCancel = nil
	}
}
