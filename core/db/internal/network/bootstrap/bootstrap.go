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

	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/network"
	"github.com/libr-forum/Libr/core/db/internal/node"
	"github.com/libr-forum/Libr/core/db/internal/routing"
)

func BootstrapFromPeers(dbnodes []*models.Node, localNode *models.Node, rt *routing.RoutingTable) {
	fmt.Println("🌐 Bootstrapping from peers...")
	for _, n := range dbnodes {
		output := ""

		if n.PeerId != "" {
			output += fmt.Sprintf("PeerId: %s", n.PeerId)
		}

		if len(n.NodeId) > 0 {
			if output != "" {
				output += ", "
			}
			output += fmt.Sprintf("NodeId: %s", base64.StdEncoding.EncodeToString(n.NodeId[:]))
		}

		if output != "" {
			fmt.Println(output)
		}
	}

	var wg sync.WaitGroup
	seen := make(map[string]bool)
	var mu sync.Mutex // protect access to `seen` map

	for _, dbnode := range dbnodes {
		// Always generate NodeId from public_key for deduplication
		if dbnode.PeerId == "" || dbnode.NodeId == [20]byte{} {
			continue
		} else if dbnode.NodeId == localNode.NodeId {
			continue
		}
		nodeID := dbnode.NodeId
		idStr := fmt.Sprintf("%x", nodeID[:])

		mu.Lock()
		if seen[idStr] {
			mu.Unlock()
			continue
		}
		seen[idStr] = true
		mu.Unlock()

		wg.Add(1)
		go func(node *models.Node) {
			defer wg.Done()
			fmt.Printf("🌐 Bootstrapping from : %s\n", idStr)
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

	if err := network.SendPing(localNode.PeerId, bootstrapNode); err != nil {
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
					"peer_id":      localNode.PeerId,
					"node_id":      base64.StdEncoding.EncodeToString(localNode.NodeId[:]),
					"find_node_id": base64.StdEncoding.EncodeToString(localNode.NodeId[:]), // Use localNode.NodeId for self
				}
				jsonBytes, _ := json.Marshal(jsonMap)

				resp, err := network.GlobalPostFunc(n.PeerId, "/route=find_node", jsonBytes)
				if err != nil {
					fmt.Printf("⚠ FindNode failed: %v (PeerId: %s)\n", err, n.PeerId)
					return
				}
				var newNodes []*models.Node
				if err := json.Unmarshal(resp, &newNodes); err != nil {
					fmt.Println("resp", resp)
					fmt.Println("Peer ID:", n.PeerId)
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
			currentClosest = append(currentClosest, base64.StdEncoding.EncodeToString(pq[i].N.NodeId[:]))
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
		rt.InsertNode(localNode, n, pinger)
		routing.GlobalRT = rt // Update the global reference

		fmt.Println("Routing Table mmm = ", rt.String())
	}
	seenMu.Unlock()

	fmt.Printf("✅ Bootstrapped with recursive lookup from %s. %d nodes added to routing table.\n",
		bootstrapNode.PeerId, len(seen))
}

func NodeUpdate(localNode *models.Node, rt *routing.RoutingTable, bootstrapAddrs []*models.Node) {
	fmt.Println("🔄 Node Update started...")

	success := false // track if any peer responds

	for _, bucket := range rt.Buckets {
		for _, dbnode := range bucket.Nodes {
			if dbnode.PeerId == "" || dbnode.NodeId == [20]byte{} {
				continue
			}
			if network.GlobalPostFunc == nil {
				fmt.Println("❌ POST function not registered in network")
				return
			}

			nodeIDStr := node.GenerateNodeIDFromPublicKey()
			jsonMap := map[string]string{
				"node_id": nodeIDStr,
				"peer_id": localNode.PeerId,
			}
			jsonBytes, _ := json.Marshal(jsonMap)

			resp, err := network.GlobalPostFunc(dbnode.PeerId, "/route=ping", jsonBytes)
			if err != nil {
				fmt.Printf("⚠ Failed to ping node %s: %v\n", dbnode.PeerId, err)
				continue
			}
			if len(resp) == 0 {
				fmt.Printf("⚠ Ping to node %s returned empty response\n", dbnode.PeerId)
				continue
			}

			fmt.Printf("✅ Node %s responded: %s\n", dbnode.PeerId, string(resp))
			success = true
		}
	}

	if !success {
		fmt.Println("❗ No existing peers responded, falling back to bootstrap nodes...")
		BootstrapFromPeers(bootstrapAddrs, localNode, rt)
	}
}
