package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	cache "github.com/libr-forum/Libr/core/mod_client/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/identity"
	"github.com/libr-forum/Libr/core/mod_client/network"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func ManualSendToMods(cert types.MsgCert, mods []types.Mod, reason string, firstTry bool) []types.ModCert {
	var (
		totalMods    = len(mods)
		ackCount     int
		rejCount     int
		unresponsive int

		modcertList []types.ModCert
		ackMods     []string // ✅ for AwaitingMods
		mu          sync.Mutex
		wg          sync.WaitGroup
	)

	if len(mods) == 0 {
		return nil
	}

	// Attach the reason (first try may have a reason, retries usually "")
	if reason != "" {
		cert.Reason = reason
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mod := range mods {
		wg.Add(1)
		go func(mod types.Mod) {
			defer wg.Done()

			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
			defer modCancel()

			respChan := make(chan interface{}, 1)

			// Send report to mod
			go func() {
				resp, err := network.SendTo(mod.PeerId, "/route=manual", cert, "mod")
				if err != nil {
					log.Printf("Error sending to %s — %v", mod.PeerId, err)
					return
				}
				respChan <- resp
			}()

			select {
			case <-modCtx.Done():
				log.Printf("Mod %s unresponsive (timeout)", mod.PeerId)
				mu.Lock()
				unresponsive++
				mu.Unlock()

			case res := <-respChan:
				modcert, ok := res.(types.ModCert)
				if !ok {
					log.Printf("Unknown response type from %s", mod.PeerId)
					return
				}

				// If they ACK, store for retry
				if modcert.Status == "acknowledged" && modcert.Sign == cert.Sign {
					mu.Lock()
					ackMods = append(ackMods, mod.PublicKey) // always store for AwaitingMods
					if firstTry {
						ackCount++ // ✅ Only count ACKs in the first try
					}
					mu.Unlock()
					log.Printf("Mod %s acknowledged", mod.PeerId)
					return
				} else {
					// Verify signature for non-acknowledgement.
					// manual_mod (image messages) are signed over content+ts+status,
					// matching the auto-mod structure. All other manual types (e.g. report)
					// use sign+status.
					var msgHash string
					if cert.Type == "manual_mod" {
						msgHash = cert.Msg.Content + strconv.FormatInt(cert.Msg.Ts, 10) + modcert.Status
					} else {
						msgHash = cert.Sign + modcert.Status
					}
					if cryptoutils.VerifySignature(modcert.PublicKey, msgHash, modcert.Sign) {
						log.Printf("Received valid modcert from %s", mod.PeerId)
						mu.Lock()
						modcertList = append(modcertList, modcert)
						if modcert.Status != "1" {
							rejCount++
						}
						mu.Unlock()
					} else {
						log.Printf("Invalid signature from mod %s", mod.PeerId)
					}
				}
			}
		}(mod)
	}

	wg.Wait()

	if firstTry {
		log.Printf("Moderation summary for %s: finalCerts=%d acks=%d unresponsive=%d total=%d",
			cert.Sign, len(modcertList), ackCount, unresponsive, totalMods)
	}

	// Save pending state only on first try — retries are persisted by the cron after
	// merging results, preserving the original AckCount.
	if len(ackMods) > 0 && firstTry {
		log.Printf("🔄 Saving %d ACK mods for retry", len(ackMods))
		signerIdentityID, err := identity.GetActiveIdentityID()
		if err != nil {
			log.Printf("Failed to read active identity for pending moderation: %v", err)
		}
		pending := types.PendingModeration{
			MsgSign:          cert.Sign,
			MsgCert:          cert,
			PartialCerts:     modcertList,
			AwaitingMods:     ackMods,
			AckCount:         len(ackMods), // needed by cron for approval ratio; prevents float div-by-zero
			SignerIdentityID: signerIdentityID,
			CreatedAt:        time.Now(),
		}

		fmt.Printf("saving pending moderation")
		if err := cache.SavePendingModeration(pending); err != nil {
			log.Printf("❌ Failed to save pending moderation: %v", err)
		} else {
			if WailsCtx != nil {
				if _, pendingQueue, err := buildPendingQueueSnapshot(); err == nil {
					runtime.EventsEmit(WailsCtx, "cron_status_update", pendingQueue)
				} else {
					log.Printf("Failed to emit pending moderation snapshot: %v", err)
				}
			}
			// StartModerationCron is idempotent — safe to call every time.
			StartModerationCron()
		}
	}

	if len(modcertList) == 0 && len(ackMods) > 0 {
		return []types.ModCert{}
	}

	return modcertList
}

func AutoSendToMods(message string, ts int64) ([]types.ModCert, error) {

	msg := types.Msg{
		Content: message,
		Ts:      ts,
	}

	onlineMods, err := util.GetOnlineMods()
	if err != nil {
		return nil, fmt.Errorf("failed to get online mods: %w", err)
	}
	if len(onlineMods) == 0 {
		return nil, fmt.Errorf("no online moderators available")
	}
	noOfMods := len(onlineMods)

	var (
		totalMods   = noOfMods
		modcertList []types.ModCert
		accpCount   int
		mu          sync.Mutex
		wg          sync.WaitGroup
		once        sync.Once
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mod := range onlineMods {
		wg.Add(1)
		go func(mod types.Mod) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] Recovered in mod goroutine for %s: %v", mod.PeerId, r)
				}
			}()

			modCtx, modCancel := context.WithTimeout(ctx, 5*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			// Send the request to the mod
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC] Recovered in mod response goroutine for %s: %v", mod.PeerId, r)
					}
				}()
				response, err := network.SendTo(mod.PeerId, "/route=auto", msg, "mod")
				log.Printf("[DEBUG] Sent to mod %s, response: %v, err: %v", mod.PeerId, response, err)
				if err != nil {
					log.Printf("[ERROR] Failed to contact mod at %s: %v", mod.PeerId, err)
					return
				}

				modcert, ok := response.(types.ModCert)
				log.Printf("[DEBUG] Modcert from %s: %v (ok=%v)", mod.PeerId, modcert, ok)
				if !ok {
					log.Printf("[ERROR] Invalid mod response format from %s: %v", mod.PeerId, response)
					return
				}

				if modcert.PublicKey != mod.PublicKey {
					log.Printf("[ERROR] Response public key mismatch from mod %s. Expected: %s, Got: %s", mod.PeerId, mod.PublicKey, modcert.PublicKey)
					return
				}

				if cryptoutils.VerifySignature(modcert.PublicKey, message+strconv.FormatInt(ts, 10)+modcert.Status, modcert.Sign) {
					log.Printf("[INFO] Valid signature from mod %s, status: %s", mod.PeerId, modcert.Status)
					responseChan <- modcert
				} else {
					log.Printf("[ERROR] Invalid signature from mod %s. Data: %s, Sign: %s", mod.PeerId, message+strconv.FormatInt(ts, 10)+modcert.Status, modcert.Sign)
				}
			}()

			select {
			case res := <-responseChan:
				if res.Status == "1" || res.Status == "0" {
					mu.Lock()
					modcertList = append(modcertList, res)
					curTotal := totalMods
					mu.Unlock()
					log.Printf("[INFO] Received modcert from %s, status: %s", mod.PeerId, res.Status)
					if res.Status == "1" {
						mu.Lock()
						accpCount++
						curAccp := accpCount
						mu.Unlock()

						log.Printf("[WARN] Mod %s Accepted. AccCount: %d, TotalMods: %d", mod.PeerId, curAccp, curTotal)
						if curAccp > (noOfMods / 2) {
							once.Do(func() {
								log.Println("Majority accepted.")
								cancel()
							})
						}
					}
				}

			case <-modCtx.Done():
				log.Printf("[WARN] Mod %s timed out or cancelled", mod.PeerId)
				mu.Lock()
				totalMods--
				curAcc := accpCount
				curTotal := totalMods
				mu.Unlock()

				log.Printf("[WARN] Timeout. RejCount: %d, TotalMods: %d", curAcc, curTotal)
				if curAcc > (noOfMods / 2) {
					once.Do(func() {
						log.Println("Majority Accepted.")
						cancel()
					})
				}
			}
		}(mod)
	}

	wg.Wait()

	mu.Lock()
	finalAccp := accpCount
	finalTotal := totalMods
	mu.Unlock()

	if finalTotal > (noOfMods/2) && float32(finalAccp)/float32(noOfMods) >= 0.3 && float32(finalAccp)/float32(finalTotal) >= 0.5 {
		return modcertList, nil
	}
	return nil, nil
}
