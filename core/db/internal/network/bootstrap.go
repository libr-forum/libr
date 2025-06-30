package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func Bootstrap(addr string, localNode *node.Node, rt *routing.RoutingTable) {
	// 1. Ping
	resp, err := http.Post("http://"+addr+"/ping", "application/json", strings.NewReader(
		fmt.Sprintf(`{"node_id": "%x"}`, localNode.NodeId[:]),
	))
	if err != nil {
		fmt.Println("Ping failed:", err)
		return
	}
	defer resp.Body.Close()

	// 2. Find Node (to populate routing table)
	findURL := "http://" + addr + "/find_node?id=" + fmt.Sprintf("%x", localNode.NodeId[:])
	resp2, err := http.Get(findURL)
	if err != nil {
		fmt.Println("FindNode failed:", err)
		return
	}
	defer resp2.Body.Close()

	var nodes []*node.Node
	if err := json.NewDecoder(resp2.Body).Decode(&nodes); err != nil {
		fmt.Println("Decoding nodes failed:", err)
		return
	}

	// 3. Insert into routing table
	pinger := &RealPinger{}
	for _, n := range nodes {
		rt.InsertNode(n, pinger)
	}
	fmt.Printf("Bootstrapped from %s. %d nodes added.\n", addr, len(nodes))
}
