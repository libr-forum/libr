// relay.go
package main

import (
	"fmt"

	"bufio"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
)

func main() {
	relayHost, err := libp2p.New(
		libp2p.EnableRelayService(),
	)

	if err != nil {
		panic(err)
	}

	autonat.New(relayHost)

	fmt.Println("[RELAY] Relay started. Listen addresses:")
	for _, addr := range relayHost.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, relayHost.ID())
	}

	relayHost.SetStreamHandler("/address-exchange/1.0.0", func(s network.Stream) {
		handleAddressExchange(relayHost, s)
	})

	select {}
}

func handleAddressExchange(h host.Host, s network.Stream) {
	defer s.Close()

	reader := bufio.NewReader(s)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("[RELAY ERROR] Failed to read peer ID:", err)
		return
	}
	line = strings.TrimSpace(line)
	targetPeerID, err := peer.Decode(line)
	if err != nil {
		fmt.Println("[RELAY ERROR] Invalid peer ID:", err)
		return
	}

	addrs := h.Peerstore().Addrs(targetPeerID)
	if len(addrs) == 0 {
		fmt.Println("[RELAY] Peer not found:", targetPeerID)
		s.Write([]byte("END\n"))
		return
	}

	fmt.Println("[RELAY] Address exchange request for:", targetPeerID)
	for _, addr := range addrs {
		s.Write([]byte(addr.String() + "\n"))
	}
	s.Write([]byte("END\n"))
}
