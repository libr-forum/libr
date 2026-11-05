package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libr-forum/Libr/core/db/internal/keycache"
	peer "github.com/libr-forum/Libr/core/db/internal/network/peers"
	"github.com/libr-forum/Libr/core/db/internal/utils"
)

func main() {
	keycache.InitKeys()

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "https://libr-relay007-1.onrender.com"
	}
	utils.InitServerClient(serverURL)

	relayAddrs, err := utils.GetRelayAddr()
	if err != nil || len(relayAddrs) == 0 {
		fmt.Println("Error while getting relay address, ", err)
		os.Exit(1)
	}

	peer.StartNode(relayAddrs)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Interrupt received. Exiting gracefully.")

	pubKeyB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
	if err := utils.DeregisterAsNode(pubKeyB64, keycache.PrivKey); err != nil {
		fmt.Println("⚠️  Failed to deregister node:", err)
	} else {
		fmt.Println("✅ Node deregistered from discovery server")
	}

	if peer.GlobalRT != nil {
		peer.GlobalRT.SaveToDBAsync()
		time.Sleep(1 * time.Second)
	}
}
