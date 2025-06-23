package core

import (
	"context"
	"fmt"
	"libr/network"
	"libr/types"
	util "libr/utils"
	"log"
	"sync"
	"time"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

func SendToMods(message string, ts int64) []types.ModCert {
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
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mod := range onlineMods {
		wg.Add(1)
		go func(mod types.Mod) {
			defer wg.Done()

			addr := fmt.Sprintf("%s:%s", mod.IP, mod.Port)
			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			// Send the request to the mod
			go func() {
				response, err := network.SendTo(addr, msg, "mod")
				if err != nil {
					log.Printf("Failed to contact mod at %s: %v", addr, err)
					return
				}

				modcert, ok := response.(types.ModCert)
				if !ok {
					log.Printf("Invalid mod response format from %s", addr)
					return
				}

				if string(modcert.PublicKey) != string(mod.PublicKey) {
					log.Printf("Response public key mismatch from mod %s", mod.IP)
					return
				}

				msgString, _ := util.CanonicalizeMsg(msg)
				if cryptoutils.VerifySignature(modcert.PublicKey, msgString, modcert.Sign) {
					responseChan <- modcert
				} else {
					log.Printf("Invalid signature from mod %s", mod.IP)
				}
			}()

			select {
			case res := <-responseChan:
				if res.Status == "approved" {
					mu.Lock()
					modcertList = append(modcertList, res)
					mu.Unlock()
				} else {
					mu.Lock()
					rejCount++
					curRej := rejCount
					curTotal := totalMods
					mu.Unlock()

					if curRej > (curTotal / 2) {
						log.Println("🚫 Majority rejected — cancelling.")
						cancel()
					}
				}

			case <-modCtx.Done():
				log.Printf("Mod %s timed out or cancelled", mod.IP)
				mu.Lock()
				totalMods--
				curRej := rejCount
				curTotal := totalMods
				mu.Unlock()

				if curRej > (curTotal / 2) {
					log.Println("🚫 Majority rejected after timeouts — cancelling.")
					cancel()
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
