package bootstrap

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func Bootstrap(targetIP string, targetPort string, localNode *node.Node, rt *routing.RoutingTable) {
	if network.GlobalPostFunc == nil {
		fmt.Println("❌ POST function not registered in network")
		return
	}

	// 1. Ping the bootstrap node
	bootstrapNode := node.Node{
		NodeId: node.GenerateNodeID(targetIP + ":" + targetPort),
		IP:     targetIP,
		Port:   targetPort,
	}
	if err := network.SendPing(localNode.NodeId, localNode.Port, bootstrapNode); err != nil {
		fmt.Println("❌ Ping failed:", err)
		return
	}
	fmt.Println("✅ Ping successful to bootstrap node")

	// 2. Prepare to perform recursive FindNode for routing table population
	targetNodeID := localNode.NodeId
	queried := make(map[string]bool)
	seen := make(map[string]*node.Node)

	var queriedMu sync.Mutex
	var seenMu sync.Mutex

	// Priority queue sorted by distance to target
	type distNode struct {
		N        *node.Node
		Distance *big.Int
	}
	var pq []distNode

	// Helper: Add node to queue
	addNode := func(n *node.Node) {
		idStr := hex.EncodeToString(n.NodeId[:])

		seenMu.Lock()
		if _, ok := seen[idStr]; ok {
			seenMu.Unlock()
			return
		}
		seen[idStr] = n
		seenMu.Unlock()

		d := node.XORBigInt(targetNodeID, n.NodeId)
		pq = append(pq, distNode{N: n, Distance: d})
	}

	// Seed with the bootstrap node
	addNode(&bootstrapNode)

	maxRounds := 3
	sameClosestCount := 0
	var lastClosest []string

	for round := 0; round < 10; round++ {
		// Sort queue by distance
		sort.Slice(pq, func(i, j int) bool {
			return pq[i].Distance.Cmp(pq[j].Distance) == -1
		})

		// Select up to alpha unqueried nodes
		var toQuery []distNode
		count := 0
		for _, dn := range pq {
			idStr := hex.EncodeToString(dn.N.NodeId[:])

			queriedMu.Lock()
			alreadyQueried := queried[idStr]
			queriedMu.Unlock()

			if !alreadyQueried {
				toQuery = append(toQuery, dn)
				count++
				if count == 3 { // alpha = 3
					break
				}
			}
		}
		if len(toQuery) == 0 {
			break
		}

		// Query them in parallel
		var wg sync.WaitGroup
		results := make(chan []*node.Node, len(toQuery))
		for _, dn := range toQuery {
			wg.Add(1)
			go func(n *node.Node) {
				defer wg.Done()
				idStr := hex.EncodeToString(n.NodeId[:])

				queriedMu.Lock()
				queried[idStr] = true
				queriedMu.Unlock()

				req := fmt.Sprintf(`{"node_id": "%x"}`, targetNodeID[:])
				resp, err := network.GlobalPostFunc(n.IP, n.Port, "/route=find_node", []byte(req))
				if err != nil {
					fmt.Println("⚠️ FindNode failed:", err)
					return
				}
				var newNodes []*node.Node
				if err := json.Unmarshal(resp, &newNodes); err != nil {
					fmt.Println("⚠️ Failed to decode FindNode response:", err)
					return
				}
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
		targetIP, targetPort, len(seen))
}
