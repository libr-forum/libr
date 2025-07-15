package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/devlup-labs/Libr/core/client/network"
	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
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
		once        sync.Once
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mod := range onlineMods {
		wg.Add(1)
		go func(mod types.Mod) {
			defer wg.Done()

			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			// Send the request to the mod
			go func() {
				response, err := network.SendTo(mod.IP, mod.Port, "/route=submit", msg, "mod")
				fmt.Println("Response:", response)
				if err != nil {
					log.Printf("Failed to contact mod at %s:%s: %v", mod.IP, mod.Port, err)
					return
				}

				modcert, ok := response.(types.ModCert)
				fmt.Println("Modcert:", modcert)
				if !ok {
					log.Printf("Invalid mod response format from %s:%s", mod.IP, mod.Port)
					return
				}

				if modcert.PublicKey != mod.PublicKey {
					log.Printf("Response public key mismatch from mod %s:%s", mod.IP, mod.Port)
					return
				}

				if cryptoutils.VerifySignature(modcert.PublicKey, msg.Content+strconv.FormatInt(msg.Ts, 10)+modcert.Status, modcert.Sign) {
					responseChan <- modcert
				} else {
					log.Printf("Invalid signature from mod %s:%s", mod.IP, mod.Port)
				}
			}()

			select {
			case res := <-responseChan:
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

					if curRej > (curTotal / 2) {
						once.Do(func() {
							log.Println("🚫 Majority rejected — cancelling.")
							cancel()
						})
					}
				}

			case <-modCtx.Done():
				log.Printf("Mod %s:%s timed out or cancelled", mod.IP, mod.Port)
				mu.Lock()
				totalMods--
				curRej := rejCount
				curTotal := totalMods
				mu.Unlock()

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
