package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/server"
)

var address string
var ip string
var port string

func main() {
	config.InitConnection()

	url := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/main/docs/database_ips.csv"
	address, err := GetFirstValidAddress(url)
	if err != nil {
		log.Fatalf("error getting address: %v", err)
	}
	fmt.Println("Using address:", address)

	id := node.GenerateNodeID(address)
	fmt.Println(hex.EncodeToString(id[:]))

	localNode := &node.Node{
		NodeId: node.GenerateNodeID(address),
		IP:     ip,
		Port:   port,
	}

	rt := routing.GetOrCreateRoutingTable(localNode)
	fmt.Println("Routing table created with port:", rt.SelfPort)

	// Optional: Bootstrap to known node
	bootstrapAddrs := os.Getenv("BOOTSTRAP")
	if bootstrapAddrs != "" {
		addrs := strings.Split(bootstrapAddrs, ",")
		for _, addr := range addrs {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				fmt.Println("Bootstrapping with", addr)
				network.Bootstrap(addr, localNode, rt)
			}
		}
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

func GetFirstValidAddress(csvURL string) (string, error) {
	resp, err := http.Get(csvURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("skipping bad row: %v", err)
			continue
		}

		if len(row) < 2 {
			log.Printf("skipping row with too few columns: %v", row)
			continue
		}

		ip := row[0]
		port := row[1]
		address := ip + ":" + port

		// TO ADD: if the node isn't working skip to next
		return address, nil
	}

	return "", fmt.Errorf("no valid address found")
}
