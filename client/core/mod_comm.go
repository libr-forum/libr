package core

import (
	"bytes"
	"context"
	"encoding/json"
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

	onlineMods, err := util.GetOnlineMods() // utils/state.go
	if err != nil {
		log.Fatalf("failed to get online mods: %v", err)
	}

	totalMods := len(onlineMods)
	var modcertList []types.ModCert

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctx, cancel := context.WithCancel(timeoutCtx)
	defer timeoutCancel()
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	rejCount := 0

	for _, mod := range onlineMods {
		wg.Add(1)
		go func(mod types.Mod, msg types.Msg) {
			defer wg.Done()

			addr := fmt.Sprintf("%s:%s", mod.IP, mod.Port)
			modCtx, modCancel := context.WithTimeout(ctx, 3*time.Second)
			defer modCancel()

			responseChan := make(chan types.ModCert, 1)

			go func() {
				response, err := network.SendTo(addr, msg, "mod")
				modcert := response.(types.ModCert)
				if !bytes.Equal(modcert.PublicKey, mod.PublicKey) {
					log.Printf("Response public key mismatch from mod %s — expected %s, got %s",
						mod.IP, mod.PublicKey, modcert.PublicKey)
					return
				}
				if err == nil {
					jsonBytes, err := json.Marshal(msg)
					if err != nil {
						log.Printf("Failed to marshal mod response from %s: %v", mod.IP, err)
						return
					}
					msgString := string(jsonBytes)
					if cryptoutils.VerifySignature(modcert.PublicKey, msgString, modcert.Sign) {
						responseChan <- modcert
					}

				}
			}()

			select {
			case res := <-responseChan:
				mu.Lock()
				if res.Status == "approved" {
					modcertList = append(modcertList, res)
				} else {
					rejCount++
					if rejCount > (totalMods / 2) {
						log.Println("Too many rejections! Cancelling all mod requests.")
						cancel()
					}
				}
				mu.Unlock()

			case <-modCtx.Done():
				log.Printf("Mod %s timed out or cancelled\n", mod.IP)
				mu.Lock()
				totalMods--
				mu.Unlock()
			}
		}(mod, msg)
	}

	wg.Wait()

	mu.Lock()
	rejected := rejCount > (totalMods / 2)
	mu.Unlock()

	if rejected {
		return nil
	}
	return modcertList
}
