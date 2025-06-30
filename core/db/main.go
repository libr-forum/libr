package main

import (
	"fmt"

	"net/http"
	"os"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables (like BOOTSTRAP_ADDR)
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found.")
	}

	// Initialize PostgreSQL connection
	config.InitConnection()

	const rtFile = "routing_table.json"
	ip := "127.0.0.1"
	port := "8001"

	// Create local node
	localNode := &node.Node{
		NodeId: node.GenerateNodeID(ip + ":" + port),
		IP:     ip,
		Port:   port,
	}

	// Load or create routing table
	rt, err := routing.LoadRoutingTable(rtFile)
	if err != nil {
		fmt.Println("No existing routing table found. Creating new.")
		rt = routing.NewRoutingTable(localNode.NodeId)
	} else {
		fmt.Println("Loaded existing routing table.")
	}

	// Bootstrap if another node is provided
	bootstrapAddr := os.Getenv("BOOTSTRAP_ADDR")
	if bootstrapAddr != "" && bootstrapAddr != ip+":"+port {
		fmt.Println("Bootstrapping from:", bootstrapAddr)
		network.Bootstrap(bootstrapAddr, localNode, rt)
	}

	// Setup all HTTP routes
	server.SetupRoutes(localNode, rt)

	fmt.Printf("✅ Node running at http://%s:%s\n", ip, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("❌ Server error:", err)
	}
}
