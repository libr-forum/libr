package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/devlup-labs/Libr/core/db/internal/network/bootstrap"
	peer "github.com/devlup-labs/Libr/core/db/internal/network/peers"
)

func main() {
	relayAddrs, err := bootstrap.GetRelayAddrs()
	if err != nil {
		fmt.Println("Error while getting relay address, ", err)
	}
	fmt.Println(relayAddrs)

	var connected bool
	for _, relayAddr := range relayAddrs {
		fmt.Println("Trying to connect to relay:", relayAddr)
		connected = peer.StartNode(relayAddr)
		if connected {
			break
		}
	}

	if !connected {
		fmt.Println("❌ Could not connect to any relay. Exiting.")
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Interrupt received. Exiting gracefully.")
}
