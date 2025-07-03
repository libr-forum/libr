package main

import (
	signalling "chatprotocol/signallingServer"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	fmt.Println("[DEBUG] Starting relay node...")

	// Create connection manager
	fmt.Println("[DEBUG] Creating connection manager...")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create connection manager: %v", err)
	}

	// Create the relay host
	fmt.Println("[DEBUG] Creating relay host...")
	relayHost, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
	)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create relay host: %v", err)
	}
	defer func() {
		fmt.Println("[DEBUG] Closing relay host...")
		relayHost.Close()
	}()

	// Enable circuit relay service
	fmt.Println("[DEBUG] Enabling circuit relay service...")
	_, err = relay.New(relayHost)
	if err != nil {
		log.Fatalf("[ERROR] Failed to enable relay service: %v", err)
	}

	fmt.Printf("[INFO] Relay started!\n")
	fmt.Printf("[INFO] Peer ID: %s\n", relayHost.ID())

	// Print all addresses
	for _, addr := range relayHost.Addrs() {
		fmt.Printf("[INFO] Relay Address: %s/p2p/%s\n", addr, relayHost.ID())
	}
	fmt.Println("[DEBUG]Starting the signalling server")
	signalling.StartServer(":9009") // assign a port to this server
	// Wait for interrupt signal
	fmt.Println("[DEBUG] Waiting for interrupt signal...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("[INFO] Shutting down relay...")

}
