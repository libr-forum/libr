package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	cache "github.com/libr-forum/Libr/core/mod_client/core/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/identity"
	"github.com/libr-forum/Libr/core/mod_client/network"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

// AutoSendToMods sends a message to all online moderators and collects their
// signatures. Returns the list of valid ModCerts received.
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

	var (
		modcertList []types.ModCert
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

			modCtx, modCancel := context.WithTimeout(ctx, 5*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[auto-mod] panic contacting %s: %v", mod.PeerId, r)
					}
				}()
				response, err := network.SendTo(mod.PeerId, "/route=auto", msg, "mod")
				if err != nil {
					log.Printf("[auto-mod] Failed to contact %s: %v", mod.PeerId, err)
					return
				}
				modcert, ok := response.(types.ModCert)
				if !ok {
					return
				}
				responseChan <- modcert
			}()

			select {
			case <-modCtx.Done():
				log.Printf("[auto-mod] Mod %s timed out", mod.PeerId)
			case modcert := <-responseChan:
				msgHash := msg.Content + strconv.FormatInt(msg.Ts, 10) + modcert.Status
				if cryptoutils.VerifySignature(modcert.PublicKey, msgHash, modcert.Sign) {
					mu.Lock()
					modcertList = append(modcertList, modcert)
					if modcert.Status == "1" {
						once.Do(func() {
							if len(modcertList) >= (len(onlineMods)/2 + 1) {
								cancel()
							}
						})
					}
					mu.Unlock()
				} else {
					log.Printf("[auto-mod] Invalid signature from %s", mod.PeerId)
				}
			}
		}(mod)
	}

	wg.Wait()
	return modcertList, nil
}

// ManualSendToMods sends a MsgCert to a specific list of mods (used for manual mod like reports).
// It saves the state to cache and starts the background moderation cron if necessary.
func ManualSendToMods(cert types.MsgCert, mods []types.Mod, reason string, firstTry bool) []types.ModCert {
	var (
		totalMods    = len(mods)
		ackCount     int
		rejCount     int
		unresponsive int

		modcertList []types.ModCert
		ackMods     []string
		mu          sync.Mutex
		wg          sync.WaitGroup
	)

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

			go func() {
				resp, err := network.SendTo(mod.PeerId, "/route=manual", cert, "mod")
				if err != nil {
					log.Printf("Error sending to %s - %v", mod.PeerId, err)
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

				if modcert.Status == "acknowledged" && modcert.Sign == cert.Sign {
					mu.Lock()
					ackMods = append(ackMods, mod.PublicKey)
					if firstTry {
						ackCount++
					}
					mu.Unlock()
					log.Printf("Mod %s acknowledged", mod.PeerId)
					return
				}

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
		}(mod)
	}

	wg.Wait()

	if firstTry {
		log.Printf("Moderation summary for %s: finalCerts=%d acks=%d unresponsive=%d total=%d",
			cert.Sign, len(modcertList), ackCount, unresponsive, totalMods)
	}

	if len(ackMods) > 0 && firstTry {
		log.Printf("Saving %d ACK mods for retry", len(ackMods))
		activeIdentityID, err := identity.GetActiveIdentityID()
		if err != nil {
			log.Printf("Failed to read active identity: %v", err)
			activeIdentityID = identity.MainIdentityID
		}

		pending := types.PendingModeration{
			MsgSign:          cert.Sign,
			MsgCert:          cert,
			PartialCerts:     modcertList,
			AwaitingMods:     ackMods,
			AckCount:         len(ackMods),
			SignerIdentityID: activeIdentityID,
			CreatedAt:        time.Now(),
		}

		if err := cache.SavePendingModeration(pending); err != nil {
			log.Printf("Failed to save pending moderation: %v", err)
		} else {
			StartModerationCron()
		}
	}

	if len(modcertList) == 0 && len(ackMods) > 0 {
		return []types.ModCert{}
	}

	return modcertList
}
