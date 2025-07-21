package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	peer "github.com/devlup-labs/Libr/core/db/internal/network/peers"
)

func main() {
	relayaddr := "/dns4/0.tcp.in.ngrok.io/tcp/13581/p2p/12D3KooWKteRVwyJ1eDMYYZjrsMC8TrfPmLgeikzixbj4GdUtBUA"
	peer.StartNode(relayaddr)

	sigChan := make(chan os.Signal, 1)

	// Notify on interrupt and terminate signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChan

	fmt.Println("Interrupt received. Exiting gracefully.")

}
