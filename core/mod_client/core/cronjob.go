package core

import (
	"log"
	"path/filepath"
	"strconv"
	"time"

	cache "github.com/devlup-labs/Libr/core/mod_client/cache_handler"
	"github.com/devlup-labs/Libr/core/mod_client/types"
	util "github.com/devlup-labs/Libr/core/mod_client/util"
)

var cronRunning = false

func StartModerationCron(msgcert *types.MsgCert) {
	if cronRunning {
		return
	}
	cronRunning = true
	log.Println("🚀 Starting moderation retry cron...")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		dir := filepath.Join(cache.GetCacheDir(), "pending_mods", "/*.json")
		files, err := filepath.Glob(dir)
		if err != nil {
			log.Printf("Cron check error: %v", err)
			continue
		}

		if len(files) == 0 {
			log.Println("✅ All moderations resolved — stopping cron")
			cronRunning = false
			return
		}

		// Retry pending moderations
		RetryPendingModerations(msgcert)
	}
}

func RetryPendingModerations(msgcert *types.MsgCert) {
	dir := filepath.Join(cache.GetCacheDir(), "pending_mods", "/*.json")
	files, err := filepath.Glob(dir)
	if err != nil {
		log.Printf("Failed to list pending moderation files: %v", err)
		return
	}

	// Get latest mod address book
	allMods, _ := util.GetOnlineMods() // returns []types.Mod with IP, Port, PublicKey, etc.

	for _, filePath := range files {
		pending, err := cache.LoadPendingModeration(filePath)
		if err != nil {
			log.Printf("Could not load pending file %s: %v", filePath, err)
			continue
		}

		// Match AwaitingMods to latest IP/Port from OnlineMods
		var retryMods []types.Mod
		for _, pubKey := range pending.AwaitingMods {
			for _, mod := range allMods {
				if mod.PublicKey == pubKey {
					retryMods = append(retryMods, mod)
					break
				}
			}
		}

		if len(retryMods) == 0 {
			continue
		}

		// Retry sending
		newCerts := ManualSendToMods(pending.MsgCert, retryMods, "")

		// Merge results
		allCerts := append(pending.PartialCerts, newCerts...)

		// Remove mods who sent final decision this round
		respondedSet := make(map[string]struct{})
		for _, mc := range newCerts {
			if mc.Status != "acknowledged" {
				respondedSet[mc.PublicKey] = struct{}{}
			}
		}

		var newAwaiting []string
		for _, pub := range pending.AwaitingMods {
			if _, ok := respondedSet[pub]; !ok {
				newAwaiting = append(newAwaiting, pub)
			}
		}

		// Update pending
		pending.PartialCerts = allCerts
		pending.AwaitingMods = newAwaiting

		// Count votes
		rejCount := 0
		for _, cert := range allCerts {
			if cert.Status == "0" {
				rejCount++
			}
		}
		ackCount := len(allCerts)

		// Decide final outcome
		if rejCount > ackCount/2 {
			log.Printf("Majority rejected — deleting %s", pending.MsgSign)
			cache.DeletePendingModeration(pending.MsgSign)
		} else if ackCount-rejCount-len(newAwaiting) > ackCount/2 {
			log.Printf("Majority approved — deleting %s", pending.MsgSign)
			tsmin := msgcert.Msg.Ts - (msgcert.Msg.Ts % 60)
			key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
			repCert := CreateRepCert(*msgcert, allCerts, "report")
			SendToDb(key, repCert, "/route=delete")
			cache.DeletePendingModeration(pending.MsgSign)
		} else {
			// Still waiting
			if err := cache.SavePendingModeration(pending); err != nil {
				log.Printf("Failed to update pending file: %v", err)
			}
		}
	}
}
