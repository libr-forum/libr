package network

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func Bootstrap(addr string, localNode *node.Node, rt *routing.RoutingTable) {
	// 1. Ping
	resp, err := http.Post("http://"+addr+"/ping", "application/json", strings.NewReader(
		fmt.Sprintf(`{"node_id": "%x","port": "%s"}`, localNode.NodeId[:], localNode.Port),
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

	bootstrapIP, bootstrapPort, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	bootstrapNode := node.Node{
		NodeId: node.GenerateNodeID(addr),
		IP:     bootstrapIP,
		Port:   bootstrapPort,
	}
	nodes = append(nodes, &bootstrapNode)
	// 3. Insert into routing table
	pinger := &RealPinger{}
	for _, n := range nodes {
		fmt.Println(n)
		rt.InsertNode(n, pinger)
	}
	fmt.Printf("Bootstrapped from %s. %d nodes added.\n", addr, len(nodes))
}
