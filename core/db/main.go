package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/server"
)

func main() {
	config.InitConnection()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // default
	}
	ip := "127.0.0.1"
	address := ip + ":" + port

	localNode := &node.Node{
		NodeId: node.GenerateNodeID(address),
		IP:     ip,
		Port:   port,
	}
	id := node.GenerateNodeID(address)
	fmt.Println(hex.EncodeToString(id[:]))

	rt := routing.GetOrCreateRoutingTable(localNode)
	fmt.Println("Routing table created with port:", rt.SelfPort)
	// Optional: Bootstrap to known node
	bootstrapAddr := os.Getenv("BOOTSTRAP")
	if bootstrapAddr != "" {
		fmt.Println("Bootstrapping with", bootstrapAddr)
		network.Bootstrap(bootstrapAddr, localNode, rt)
	}

	server.SetupRoutes(localNode, rt)
	data, err := json.MarshalIndent(rt, "", "  ")
	if err != nil {
		log.Println("Error marshalling routing table:", err)
		return
	}
	fmt.Println(string(data))
	fmt.Println("Kademlia node running at http://" + address)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
	fmt.Println(rt)
}
