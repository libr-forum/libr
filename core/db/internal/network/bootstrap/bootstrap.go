// package bootstrap

// import (
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"math/big"
// 	"reflect"
// 	"sort"
// 	"strings"
// 	"sync"

// 	"github.com/devlup-labs/Libr/core/db/internal/network"
// 	"github.com/devlup-labs/Libr/core/db/internal/node"
// 	"github.com/devlup-labs/Libr/core/db/internal/routing"
// )

// func BootstrapFromPeers(peers []string, localNode *models.Node, rt *routing.RoutingTable) {
// 	var wg sync.WaitGroup
// 	seen := make(map[string]bool)
// 	var mu sync.Mutex // protect access to `seen` map

// 	for _, addr := range peers {
// 		// Skip invalid addresses
// 		parts := strings.Split(addr, ":")
// 		if len(parts) != 2 {
// 			log.Printf("⚠️ Skipping invalid bootstrap address: %s", addr)
// 			continue
// 		}
// 		ip := parts[0]
// 		port := parts[1]

// 		// Deduplication check
// 		mu.Lock()
// 		if seen[addr] {
// 			mu.Unlock()
// 			continue
// 		}
// 		seen[addr] = true
// 		mu.Unlock()

// 		wg.Add(1)
// 		go func(ip, port, addr string) {
// 			defer wg.Done()
// 			fmt.Printf("🌐 Bootstrapping from %s\n", addr)
// 			Bootstrap(ip, port, localNode, rt)
// 		}(ip, port, addr)
// 	}

// 	wg.Wait()
// }

// func Bootstrap(targetIP string, targetPort string, localNode *models.Node, rt *routing.RoutingTable) {
// 	if network.GlobalPostFunc == nil {
// 		fmt.Println("❌ POST function not registered in network")
// 		return
// 	}

// 	// 1. Ping the bootstrap node
// 	bootstrapNode := models.Node{
// 		NodeId: node.GenerateNodeID(targetIP + ":" + targetPort),
// 		IP:     targetIP,
// 		Port:   targetPort,
// 	}
// 	if err := network.SendPing(localmodels.NodeId, localNode.Port, bootstrapNode); err != nil {
// 		fmt.Println("❌ Ping failed:", err)
// 		return
// 	}
// 	fmt.Println("✅ Ping successful to bootstrap node")

// 	// 2. Prepare to perform recursive FindNode for routing table population
// 	targetNodeID := localmodels.NodeId
// 	queried := make(map[string]bool)
// 	seen := make(map[string]*models.Node)

// 	var queriedMu sync.Mutex
// 	var seenMu sync.Mutex

// 	// Priority queue sorted by distance to target
// 	type distNode struct {
// 		N        *models.Node
// 		Distance *big.Int
// 	}
// 	var pq []distNode

// 	// Helper: Add node to queue
// 	addNode := func(n *models.Node) {
// 		idStr := hex.EncodeToString(n.NodeId[:])

// 		seenMu.Lock()
// 		if _, ok := seen[idStr]; ok {
// 			seenMu.Unlock()
// 			return
// 		}
// 		seen[idStr] = n
// 		seenMu.Unlock()

// 		d := node.XORBigInt(targetNodeID, n.NodeId)
// 		pq = append(pq, distNode{N: n, Distance: d})
// 	}

// 	// Seed with the bootstrap node
// 	addNode(&bootstrapNode)

// 	maxRounds := 3
// 	sameClosestCount := 0
// 	var lastClosest []string

// 	for round := 0; round < 10; round++ {
// 		// Sort queue by distance
// 		sort.Slice(pq, func(i, j int) bool {
// 			return pq[i].Distance.Cmp(pq[j].Distance) == -1
// 		})

// 		// Select up to alpha unqueried nodes
// 		var toQuery []distNode
// 		count := 0
// 		for _, dn := range pq {
// 			idStr := hex.EncodeToString(dn.N.NodeId[:])

// 			queriedMu.Lock()
// 			alreadyQueried := queried[idStr]
// 			queriedMu.Unlock()

// 			if !alreadyQueried {
// 				toQuery = append(toQuery, dn)
// 				count++
// 				if count == 3 { // alpha = 3
// 					break
// 				}
// 			}
// 		}
// 		if len(toQuery) == 0 {
// 			break
// 		}

// 		// Query them in parallel
// 		var wg sync.WaitGroup
// 		results := make(chan []*models.Node, len(toQuery))
// 		for _, dn := range toQuery {
// 			wg.Add(1)
// 			go func(n *models.Node) {
// 				defer wg.Done()
// 				idStr := hex.EncodeToString(n.NodeId[:])

// 				queriedMu.Lock()
// 				queried[idStr] = true
// 				queriedMu.Unlock()

// 				req := fmt.Sprintf(`{"node_id": "%x"}`, targetNodeID[:])
// 				resp, err := network.GlobalPostFunc(n.IP, n.Port, "/route=find_node", []byte(req))
// 				if err != nil {
// 					fmt.Println("⚠️ FindNode failed:", err)
// 					return
// 				}
// 				var newNodes []*models.Node
// 				if err := json.Unmarshal(resp, &newNodes); err != nil {
// 					fmt.Println("⚠️ Failed to decode FindNode response:", err)
// 					return
// 				}
// 				results <- newNodes
// 			}(dn.N)
// 		}
// 		wg.Wait()
// 		close(results)

// 		for res := range results {
// 			for _, n := range res {
// 				addNode(n)
// 			}
// 		}

// 		// Convergence check
// 		currentClosest := []string{}
// 		for i := 0; i < len(pq) && i < 20; i++ {
// 			currentClosest = append(currentClosest, hex.EncodeToString(pq[i].N.NodeId[:]))
// 		}
// 		if reflect.DeepEqual(currentClosest, lastClosest) {
// 			sameClosestCount++
// 		} else {
// 			sameClosestCount = 0
// 		}
// 		lastClosest = currentClosest

// 		if sameClosestCount >= maxRounds {
// 			fmt.Println("✅ Converged after", round+1, "rounds.")
// 			break
// 		}
// 	}

// 	// Add all seen nodes to routing table
// 	pinger := &network.RealPinger{}
// 	seenMu.Lock()
// 	for _, n := range seen {
// 		rt.InsertNode(n, pinger)
// 	}
// 	seenMu.Unlock()

// 	fmt.Printf("✅ Bootstrapped with recursive lookup from %s:%s. %d nodes added to routing table.\n",
// 		targetIP, targetPort, len(seen))
// }

package bootstrap

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func BootstrapFromPeers(dbnodes []*models.Node, localNode *models.Node, rt *routing.RoutingTable) {
	fmt.Println("dbnodes:")
	for _, n := range dbnodes {
		fmt.Printf(n.IP, n.Port, n.NodeId, n.PublicKey)
	}
	var wg sync.WaitGroup
	seen := make(map[string]bool)
	var mu sync.Mutex // protect access to `seen` map

	for _, dbnode := range dbnodes {
		// Always generate NodeId from public_key for deduplication
		if dbnode.PublicKey == "" {
			continue
		}
		nodeID := node.GenerateNodeID(dbnode.PublicKey)
		idStr := fmt.Sprintf("%x", nodeID[:])

		mu.Lock()
		if seen[idStr] {
			mu.Unlock()
			continue
		}
		seen[idStr] = true
		mu.Unlock()

		// Ensure NodeId is set correctly
		dbnode.NodeId = nodeID

		wg.Add(1)
		go func(node *models.Node) {
			defer wg.Done()
			fmt.Printf("🌐 Bootstrapping from %s:%s\n", node.IP, node.Port)
			Bootstrap(node, localNode, rt)
		}(dbnode)
	}

	wg.Wait()
}

func Bootstrap(bootstrapNode *models.Node, localNode *models.Node, rt *routing.RoutingTable) {
	if network.GlobalPostFunc == nil {
		fmt.Println("❌ POST function not registered in network")
		return
	}

	if err := network.SendPing(localNode.NodeId, localNode.Port, bootstrapNode); err != nil {
		fmt.Println("❌ Ping failed:", err)
		return
	}
	fmt.Println("✅ Ping successful to bootstrap node")

	// 2. Prepare to perform recursive FindNode for routing table population
	queried := make(map[string]bool)
	seen := make(map[string]*models.Node)

	var queriedMu sync.Mutex
	var seenMu sync.Mutex

	// Priority queue sorted by distance to target
	type distNode struct {
		N        *models.Node
		Distance *big.Int
	}
	var pq []distNode

	// Helper: Add node to queue (deduplicated and mark as queried immediately)
	addNode := func(n *models.Node) {
		idStr := hex.EncodeToString(n.NodeId[:])
		seenMu.Lock()
		defer seenMu.Unlock()
		if _, ok := seen[idStr]; ok {
			return
		}
		// Do NOT mark as queried here!
		// Only add to seen
		seen[idStr] = n
		d := node.XORBigInt(localNode.NodeId, n.NodeId)
		pq = append(pq, distNode{N: n, Distance: d})
	}

	// Seed with the bootstrap node
	addNode(bootstrapNode)

	maxRounds := 3
	sameClosestCount := 0
	var lastClosest []string

	for round := 0; round < 10; round++ {
		// Sort queue by distance
		sort.Slice(pq, func(i, j int) bool {
			return pq[i].Distance.Cmp(pq[j].Distance) == -1
		})

		// Select up to alpha unqueried nodes, and mark as queried immediately
		var toQuery []distNode
		count := 0
		for _, dn := range pq {
			idStr := hex.EncodeToString(dn.N.NodeId[:])
			queriedMu.Lock()
			if queried[idStr] {
				queriedMu.Unlock()
				continue
			}
			queried[idStr] = true
			toQuery = append(toQuery, dn)
			count++
			queriedMu.Unlock()
			if count == 3 { // alpha = 3
				break
			}
		}
		if len(toQuery) == 0 {
			break
		}

		// Query them in parallel
		var wg sync.WaitGroup
		results := make(chan []*models.Node, len(toQuery))
		for _, dn := range toQuery {
			wg.Add(1)
			go func(n *models.Node) {
				defer wg.Done()
				jsonMap := map[string]string{
					"node_id":    base64.StdEncoding.EncodeToString(localNode.NodeId[:]),
					"public_key": localNode.PublicKey[:],
				}
				jsonBytes, _ := json.Marshal(jsonMap)
				fmt.Println("public_key in find_node:", localNode.PublicKey[:])
				resp, err := network.GlobalPostFunc(n.IP, n.Port, "/route=find_node", jsonBytes)
				if err != nil {
					fmt.Println("⚠ FindNode failed:", err)
					return
				}
				var newNodes []*models.Node
				if err := json.Unmarshal(resp, &newNodes); err != nil {
					fmt.Println("⚠ Failed to decode FindNode response:", err)
					return
				}
				fmt.Println("newNodes:", newNodes)
				results <- newNodes
			}(dn.N)
		}
		wg.Wait()
		close(results)

		for res := range results {
			for _, n := range res {
				addNode(n)
			}
		}

		// Convergence check
		currentClosest := []string{}
		for i := 0; i < len(pq) && i < 20; i++ {
			currentClosest = append(currentClosest, hex.EncodeToString(pq[i].N.NodeId[:]))
		}
		if reflect.DeepEqual(currentClosest, lastClosest) {
			sameClosestCount++
		} else {
			sameClosestCount = 0
		}
		lastClosest = currentClosest

		if sameClosestCount >= maxRounds {
			fmt.Println("✅ Converged after", round+1, "rounds.")
			break
		}
	}

	// Add all seen nodes to routing table
	pinger := &network.RealPinger{}
	seenMu.Lock()
	for _, n := range seen {
		rt.InsertNode(n, pinger)
	}
	seenMu.Unlock()

	fmt.Printf("✅ Bootstrapped with recursive lookup from %s:%s. %d nodes added to routing table.\n",
		bootstrapNode.IP, bootstrapNode.Port, len(seen))
}

func NodeUpdate(rt *routing.RoutingTable) {
	fmt.Println("Node Update heheheh")
	for _, bucket := range rt.Buckets {
		for _, node := range bucket.Nodes {
			if node.IP == "" || node.Port == "" {
				continue
			}
			if network.GlobalPostFunc == nil {
				fmt.Println("❌ POST function not registered in network")
				return
			}

			req := fmt.Sprintf(`{"node_id": "%x","public_key": "%x"}`, node.NodeId[:], node.PublicKey[:])
			resp, err := network.GlobalPostFunc(node.IP, node.Port, "/route=ping", []byte(req))
			if err != nil {
				fmt.Printf("⚠ Failed to ping node %s:%s: %v\n", node.IP, node.Port, err)
				continue
			}
			fmt.Printf("✅ Node %s:%s responded: %s\n", node.IP, node.Port, string(resp))
		}
	}
}
