package core

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	cache "github.com/devlup-labs/Libr/core/mod_client/cache_handler"
	"github.com/devlup-labs/Libr/core/mod_client/network"
	"github.com/devlup-labs/Libr/core/mod_client/types"
	util "github.com/devlup-labs/Libr/core/mod_client/util"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
)

// func SendToMods(message string, ts int64, reason *string, action string, sign *string) []types.ModCert {

// 	msg := types.SubmitMsg{
// 		Content: message,
// 		Ts:      ts,
// 		Reason:  reason,
// 		Mode:    action,
// 		Sign:    sign,
// 	}

// 	onlineMods, err := util.GetOnlineMods()
// 	if err != nil {
// 		log.Fatalf("failed to get online mods: %v", err)
// 	}

// 	var (
// 		totalMods   = len(onlineMods)
// 		modcertList []types.ModCert
// 		rejCount    int
// 		mu          sync.Mutex
// 		wg          sync.WaitGroup
// 		once        sync.Once
// 	)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	for _, mod := range onlineMods {
// 		wg.Add(1)
// 		go func(mod types.Mod) {
// 			defer wg.Done()

// 			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
// 			defer modCancel()

// 			responseChan := make(chan types.ModCert, 1)

// 			// Send the request to the mod
// 			go func() {
// 				response, err := network.SendTo(mod.IP, mod.Port, "/route=submit", msg, "mod")
// 				fmt.Println("Response:", response)
// 				if err != nil {
// 					log.Printf("Failed to contact mod at %s:%s: %v", mod.IP, mod.Port, err)
// 					return
// 				}

// 				switch resp := response.(type) {
// 				case types.ModCert:
// 					if resp.PublicKey != mod.PublicKey {
// 						log.Printf("Response public key mismatch from mod %s:%s", mod.IP, mod.Port)
// 						return
// 					}

// 					// Verify mod signature
// 					if cryptoutils.VerifySignature(resp.PublicKey, message+strconv.FormatInt(ts, 10)+resp.Status, resp.Sign) {
// 						responseChan <- resp
// 					} else {
// 						log.Printf("Invalid signature from mod %s:%s", mod.IP, mod.Port)
// 					}
// 				case string:
// 					if strings.ToLower(resp) == "acknowledgement" {
// 						log.Printf("Manual mode: received acknowledgement from %s:%s", mod.IP, mod.Port)
// 						// Treat as neutral or accept based on your policy
// 						// Optional: count as approve? Or ignore?
// 					} else {
// 						log.Printf("Manual mode: unexpected string response from %s:%s: %v", mod.IP, mod.Port, resp)
// 					}
// 				case nil:
// 					log.Printf("No response or nil response from mod %s:%s", mod.IP, mod.Port)
// 				default:
// 					log.Printf("Unexpected response type from mod %s:%s: %T", mod.IP, mod.Port, resp)
// 				}

// 				if modcert.PublicKey != mod.PublicKey {
// 					log.Printf("Response public key mismatch from mod %s:%s", mod.IP, mod.Port)
// 					return
// 				}

// 				if cryptoutils.VerifySignature(modcert.PublicKey, message+strconv.FormatInt(ts, 10)+modcert.Status, modcert.Sign) {
// 					responseChan <- modcert
// 				} else {
// 					log.Printf("Invalid signature from mod %s:%s", mod.IP, mod.Port)
// 				}
// 			}()

// 			select {
// 			case res := <-responseChan:
// 				if res.Status == "1" {
// 					mu.Lock()
// 					modcertList = append(modcertList, res)
// 					mu.Unlock()
// 				} else {
// 					mu.Lock()
// 					rejCount++
// 					curRej := rejCount
// 					curTotal := totalMods
// 					mu.Unlock()

// 					if curRej > (curTotal / 2) {
// 						once.Do(func() {
// 							log.Println("🚫 Majority rejected — cancelling.")
// 							cancel()
// 						})
// 					}
// 				}

// 			case <-modCtx.Done():
// 				log.Printf("Mod %s:%s timed out or cancelled", mod.IP, mod.Port)
// 				mu.Lock()
// 				totalMods--
// 				curRej := rejCount
// 				curTotal := totalMods
// 				mu.Unlock()

// 				if curRej > (curTotal / 2) {
// 					once.Do(func() {
// 						log.Println("🚫 Majority rejected after timeouts — cancelling.")
// 						cancel()
// 					})
// 				}
// 			}
// 		}(mod)
// 	}

// 	wg.Wait()

// 	mu.Lock()
// 	finalRej := rejCount
// 	finalTotal := totalMods
// 	mu.Unlock()

// 	if finalRej > (finalTotal / 2) {
// 		return nil
// 	}
// 	return modcertList
// }

// func SendToModsManual(message string, ts int64, reason *string, action string, sign *string) ManualModStats {
// 	msg := types.SubmitMsg{
// 		Content: message,
// 		Ts:      ts,
// 		Reason:  reason,
// 		Mode:    action,
// 		Sign:    sign,
// 	}

// 	onlineMods, err := util.GetOnlineMods()
// 	if err != nil {
// 		log.Fatalf("failed to get online mods: %v", err)
// 	}

// 	var (
// 		stats ManualModStats
// 		mu    sync.Mutex
// 		wg    sync.WaitGroup
// 	)

// 	for _, mod := range onlineMods {
// 		wg.Add(1)
// 		go func(mod types.Mod) {
// 			defer wg.Done()

// 			modCtx, modCancel := context.WithTimeout(context.Background(), 3*time.Second)
// 			defer modCancel()

// 			response, err := network.SendTo(mod.IP, mod.Port, "/route=submit", msg, "mod")
// 			if err != nil {
// 				mu.Lock()
// 				stats.Unresponsive++
// 				mu.Unlock()
// 				return
// 			}

// 			switch v := response.(type) {
// 			case string:
// 				if v == "Acknowledgement" {
// 					mu.Lock()
// 					stats.Acks++
// 					mu.Unlock()
// 				}
// 			case types.ModCert:
// 				// Validate sig and public key if needed
// 				if v.PublicKey == mod.PublicKey &&
// 					cryptoutils.VerifySignature(v.PublicKey, message+strconv.FormatInt(ts, 10)+v.Status, v.Sign) {
// 					mu.Lock()
// 					stats.ModCerts = append(stats.ModCerts, v)
// 					mu.Unlock()
// 				}
// 			default:
// 				mu.Lock()
// 				stats.Unresponsive++
// 				mu.Unlock()
// 			}
// 		}(mod)
// 	}

// 	wg.Wait()
// 	return stats
// }

type ManualModStats struct {
	Acks         int
	ModCerts     []types.ModCert
	Unresponsive int
}

// func ManualSendToMods(msgcert types.MsgCert) []types.ModCert {

// 	onlineMods, err := util.GetOnlineMods()
// 	if err != nil {
// 		log.Fatalf("failed to get online mods: %v", err)
// 	}

// 	var (
// 		totalMods   = len(onlineMods)
// 		modcertList []types.ModCert
// 		rejCount    int
// 		mu          sync.Mutex
// 		wg          sync.WaitGroup
// 		stats       ManualModStats
// 		once        sync.Once
// 	)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	for _, mod := range onlineMods {
// 		wg.Add(1)
// 		go func(mod types.Mod) {
// 			defer wg.Done()

// 			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
// 			defer modCancel()

// 			responseChan := make(chan types.ModCert, 1)

// 			// Send the request to the mod
// 			go func() {
// 				response, err := network.SendTo(mod.IP, mod.Port, "/route=submit", msgcert, "mod")
// 				fmt.Println("Response:", response)
// 				if err != nil {
// 					log.Printf("Failed to contact mod at %s:%s: %v", mod.IP, mod.Port, err)
// 					return
// 				}

// 				switch v := response.(type) {
// 				case string:
// 					if v == "Acknowledgement" {
// 						mu.Lock()
// 						stats.Acks++
// 						mu.Unlock()
// 					}
// 				case types.ModCert:
// 					if v.PublicKey != mod.PublicKey {
// 						log.Printf("Response public key mismatch from mod %s:%s", mod.IP, mod.Port)
// 						return
// 					}

// 					if cryptoutils.VerifySignature(v.PublicKey, msgcert.Msg.Content+strconv.FormatInt(msgcert.Msg.Ts, 10)+v.Status, v.Sign) {
// 						responseChan <- v
// 					} else {
// 						log.Printf("Invalid signature from mod %s:%s", mod.IP, mod.Port)
// 					}

// 					mu.Lock()
// 					stats.ModCerts = append(stats.ModCerts, v)
// 					mu.Unlock()

// 				default:
// 					mu.Lock()
// 					stats.Unresponsive++
// 					mu.Unlock()
// 				}

// 				fmt.Println("Modcert:", v)

// 			}()

// 			select {
// 			case res := <-responseChan:
// 				if res.Status == "1" {
// 					mu.Lock()
// 					modcertList = append(modcertList, res)
// 					mu.Unlock()
// 				} else {
// 					mu.Lock()
// 					rejCount++
// 					curRej := rejCount
// 					curTotal := totalMods
// 					mu.Unlock()

// 					if curRej > (curTotal / 2) {
// 						once.Do(func() {
// 							log.Println("🚫 Majority rejected — cancelling.")
// 							cancel()
// 						})
// 					}
// 				}

// 			case <-modCtx.Done():
// 				log.Printf("Mod %s:%s timed out or cancelled", mod.IP, mod.Port)
// 				mu.Lock()
// 				totalMods--
// 				curRej := rejCount
// 				curTotal := totalMods
// 				mu.Unlock()

// 				if curRej > (curTotal / 2) {
// 					once.Do(func() {
// 						log.Println("🚫 Majority rejected after timeouts — cancelling.")
// 						cancel()
// 					})
// 				}
// 			}
// 		}(mod)
// 	}

// 	wg.Wait()

// 	mu.Lock()
// 	finalRej := rejCount
// 	finalTotal := totalMods
// 	mu.Unlock()

// 	if finalRej > (finalTotal / 2) {
// 		return nil
// 	}
// 	return modcertList
// }

func ManualSendToMods(cert types.MsgCert, mods []types.Mod, reason string) []types.ModCert {
	var (
		totalMods    = len(mods)
		ackCount     int
		rejCount     int
		unresponsive int

		modcertList []types.ModCert
		responded   = make(map[string]bool)

		mu sync.Mutex
		wg sync.WaitGroup
	)

	// Attach the reason to cert before sending
	cert.Reason = reason

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
				resp, err := network.SendTo(mod.IP, mod.Port, "/route=report", cert, "mod")
				if err != nil {
					log.Printf("Error sending to %s:%s — %v", mod.IP, mod.Port, err)
					return
				}
				respChan <- resp
			}()

			select {
			case <-modCtx.Done():
				log.Printf("Mod %s:%s unresponsive (timeout)", mod.IP, mod.Port)
				mu.Lock()
				unresponsive++
				mu.Unlock()

			case res := <-respChan:
				mu.Lock()
				responded[mod.PublicKey] = true
				mu.Unlock()

				modcert, ok := res.(types.ModCert)
				if !ok {
					log.Printf("Unknown response type from %s:%s", mod.IP, mod.Port)
					return
				}
				if modcert.PublicKey != mod.PublicKey {
					log.Printf("Invalid mod cert from %s:%s (wrong pubkey)", mod.IP, mod.Port)
					return
				}
				if modcert.Status == "acknowledgement" && modcert.Sign == cert.Sign {
					log.Printf("Mod %s:%s acknowledged", mod.IP, mod.Port)
					mu.Lock()
					ackCount++
					mu.Unlock()
					return
				}
				// Verify signature for non-acknowledgement
				msgHash := cert.Sign + modcert.Status
				if cryptoutils.VerifySignature(modcert.PublicKey, msgHash, modcert.Sign) {
					log.Printf("Received valid modcert from %s:%s", mod.IP, mod.Port)
					mu.Lock()
					modcertList = append(modcertList, modcert)
					if modcert.Status != "1" {
						rejCount++
					}
					mu.Unlock()
				} else {
					log.Printf("Invalid signature from mod %s:%s", mod.IP, mod.Port)
				}
			}
		}(mod)
	}

	wg.Wait()

	log.Printf("Moderation summary for %s: certs=%d acks=%d unresponsive=%d total=%d",
		cert.Sign, len(modcertList), ackCount, unresponsive, totalMods)

	// If not enough certs collected, save state and launch cron job
	if len(modcertList) < 3 && len(modcertList) <= totalMods/2 {
		log.Printf("🔄 Not enough modcerts — saving and scheduling retry")
		awaitingMods := []string{}
		for _, mod := range mods {
			if !responded[mod.PublicKey] {
				awaitingMods = append(awaitingMods, mod.PublicKey)
			}
		}
		pending := types.PendingModeration{
			MsgSign:      cert.Sign,
			MsgCert:      cert,
			PartialCerts: modcertList,
			AwaitingMods: awaitingMods,
			AckCount:     ackCount,
			CreatedAt:    time.Now(),
		}

		// ✅ Persist to disk
		if err := cache.SavePendingModeration(pending); err != nil {
			log.Printf("❌ Failed to save pending moderation: %v", err)
		} else {
			go StartModerationCron(&cert)
		}
	}

	return modcertList
}

func AutoSendToMods(message string, ts int64) []types.ModCert {

	msg := types.Msg{
		Content: message,
		Ts:      ts,
	}

	onlineMods, err := util.GetOnlineMods()
	if err != nil {
		log.Fatalf("failed to get online mods: %v", err)
	}

	var (
		totalMods   = len(onlineMods)
		modcertList []types.ModCert
		rejCount    int
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
					log.Printf("[PANIC] Recovered in mod goroutine for %s:%s: %v", mod.IP, mod.Port, r)
				}
			}()

			modCtx, modCancel := context.WithTimeout(ctx, 5*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			// Send the request to the mod
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC] Recovered in mod response goroutine for %s:%s: %v", mod.IP, mod.Port, r)
					}
				}()
				response, err := network.SendTo(mod.IP, mod.Port, "/route=submit", msg, "mod")
				log.Printf("[DEBUG] Sent to mod %s:%s, response: %v, err: %v", mod.IP, mod.Port, response, err)
				if err != nil {
					log.Printf("[ERROR] Failed to contact mod at %s:%s: %v", mod.IP, mod.Port, err)
					return
				}

				modcert, ok := response.(types.ModCert)
				log.Printf("[DEBUG] Modcert from %s:%s: %v (ok=%v)", mod.IP, mod.Port, modcert, ok)
				if !ok {
					log.Printf("[ERROR] Invalid mod response format from %s:%s: %v", mod.IP, mod.Port, response)
					return
				}

				if modcert.PublicKey != mod.PublicKey {
					log.Printf("[ERROR] Response public key mismatch from mod %s:%s. Expected: %s, Got: %s", mod.IP, mod.Port, mod.PublicKey, modcert.PublicKey)
					return
				}

				if cryptoutils.VerifySignature(modcert.PublicKey, message+strconv.FormatInt(ts, 10)+modcert.Status, modcert.Sign) {
					log.Printf("[INFO] Valid signature from mod %s:%s, status: %s", mod.IP, mod.Port, modcert.Status)
					responseChan <- modcert
				} else {
					log.Printf("[ERROR] Invalid signature from mod %s:%s. Data: %s, Sign: %s", mod.IP, mod.Port, message+strconv.FormatInt(ts, 10)+modcert.Status, modcert.Sign)
				}
			}()

			select {
			case res := <-responseChan:
				log.Printf("[INFO] Received modcert from %s:%s, status: %s", mod.IP, mod.Port, res.Status)
				if res.Status == "1" {
					mu.Lock()
					modcertList = append(modcertList, res)
					mu.Unlock()
				} else {
					mu.Lock()
					rejCount++
					curRej := rejCount
					curTotal := totalMods
					mu.Unlock()

					log.Printf("[WARN] Mod %s:%s rejected. RejCount: %d, TotalMods: %d", mod.IP, mod.Port, curRej, curTotal)
					if curRej > (curTotal / 2) {
						once.Do(func() {
							log.Println("🚫 Majority rejected — cancelling.")
							cancel()
						})
					}
				}

			case <-modCtx.Done():
				log.Printf("[WARN] Mod %s:%s timed out or cancelled", mod.IP, mod.Port)
				mu.Lock()
				totalMods--
				curRej := rejCount
				curTotal := totalMods
				mu.Unlock()

				log.Printf("[WARN] Timeout. RejCount: %d, TotalMods: %d", curRej, curTotal)
				if curRej > (curTotal / 2) {
					once.Do(func() {
						log.Println("🚫 Majority rejected after timeouts — cancelling.")
						cancel()
					})
				}
			}
		}(mod)
	}

	wg.Wait()

	mu.Lock()
	finalRej := rejCount
	finalTotal := totalMods
	mu.Unlock()

	if finalRej > (finalTotal / 2) {
		return nil
	}
	return modcertList
}
