package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

type reqFormat struct {
	Type  string `json:"type"`
	PubIP string `json:"pubip"`
}

var IDmap map[string]string

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

	relayHost.SetStreamHandler("/chat/1.0.0", handleChatStream)
	// Wait for interrupt signal
	fmt.Println("[DEBUG] Waiting for interrupt signal...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("[INFO] Shutting down relay...")

}

func handleChatStream(s network.Stream) {
	fmt.Println("[DEBUG] Incoming chat stream from", s.Conn().RemoteMultiaddr())
	defer s.Close()
	reader := bufio.NewReader(s)
	for {
		var req reqFormat
		buf := new([]byte)
		_, err := reader.Read(*buf)
		if err!=nil{
			fmt.Println("[DEBUG]Error vreating reader at relay")
		}

		err = json.Unmarshal(*buf, &req)

		if err != nil {
			fmt.Printf("[DEBUG]Error getting req as json at relay")
			return
		}

		if req.Type == "register" {
			remoteAddr := s.Conn().RemoteMultiaddr()
			 peerID := s.Conn().RemotePeer()
   			// Combine to full multiaddress string
   			fullMultiAddr := fmt.Sprintf("%s/p2p/%s", remoteAddr.String(), peerID.String())
			fmt.Println("[INFO]Registering the peer into relay map")
			IDmap[req.PubIP] = fullMultiAddr
		}
		if req.Type == "getID" {
			s.Write([]byte(IDmap[req.PubIP]+"\n"))
		}
		
	}

}
