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

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		files, err := filepath.Glob("pending_mods/*.json")
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
	files, err := filepath.Glob("pending_mods/*.json")
	if err != nil {
		log.Printf("Failed to list pending moderation files: %v", err)
		return
	}

	for _, filePath := range files {
		pending, err := cache.LoadPendingModeration(filePath)
		if err != nil {
			log.Printf("Could not load pending file %s: %v", filePath, err)
			continue
		}

		// Get all online mods
		totalMods, err := util.GetOnlineMods()
		if err != nil {
			log.Printf("Failed to get online mods: %v", err)
			continue
		}

		// Filter awaiting mods from total mods
		awaitingSet := make(map[string]struct{})
		for _, pk := range pending.AwaitingMods {
			awaitingSet[pk] = struct{}{}
		}

		var awaitingMods []types.Mod
		for _, mod := range totalMods {
			if _, ok := awaitingSet[mod.PublicKey]; ok {
				awaitingMods = append(awaitingMods, mod)
			}
		}

		// Retry with remaining mods
		pending.MsgCert.ModCerts = pending.PartialCerts
		newCerts := ManualSendToMods(pending.MsgCert, awaitingMods, "")

		// Merge all certs
		allCerts := append(pending.PartialCerts, newCerts...)

		// Count approvals and acknowledgements
		rejCount := 0
		ackCount := 0
		for _, cert := range allCerts {
			if cert.Status == "0" {
				rejCount++
			}
			ackCount++ // all modcerts are acks by definition
		}

		// Save updated counts
		pending.AckCount = ackCount

		if rejCount > ackCount/2 {
			log.Printf("✅ Majority declined (%d/%d) — deleting %s", rejCount, ackCount, pending.MsgSign)
			cache.DeletePendingModeration(pending.MsgSign)
			// Optional: Finalize moderation action here
		} else if ackCount-rejCount-len(awaitingMods) > ackCount/2 {
			log.Printf("✅ Majority approved (%d/%d) — deleting %s", ackCount-rejCount-len(awaitingMods), ackCount, pending.MsgSign)
			tsmin := msgcert.Msg.Ts - (msgcert.Msg.Ts % 60)
			key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
			repCert := CreateRepCert(*msgcert, allCerts, "report")
			SendToDb(key, repCert, "/route=delete")
			cache.DeletePendingModeration(pending.MsgSign)

		} else {

			log.Printf("⏳ Still waiting — updating %s", pending.MsgSign)

			// Remove newly responded mods from awaiting list
			respondedSet := map[string]struct{}{}
			for _, mc := range newCerts {
				respondedSet[mc.PublicKey] = struct{}{}
			}
			var newAwaiting []string
			for _, modPub := range pending.AwaitingMods {
				if _, ok := respondedSet[modPub]; !ok {
					newAwaiting = append(newAwaiting, modPub)
				}
			}

			// Update and save
			pending.PartialCerts = allCerts
			pending.AwaitingMods = newAwaiting
			if err := cache.SavePendingModeration(pending); err != nil {
				log.Printf("Failed to update pending file: %v", err)
			}
		}
	}
}
