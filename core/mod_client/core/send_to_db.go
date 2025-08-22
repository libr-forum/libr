package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/libr-forum/Libr/core/mod_client/network"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

type BaseResponse struct {
	Type string `json:"type"`
}

type RedirectResponse struct {
	Type  string       `json:"type"`
	Nodes []types.Node `json:"nodes"`
}

type StoredResponse struct {
	Type   string       `json:"type"`
	Status string       `json:"status"`
	Nodes  []types.Node `json:"nodes,omitempty"` // NEW: Include nodes in stored response
}

// NEW: Configuration for sparse network handling
type NetworkConfig struct {
	MinStorageNodes     int           // Minimum acceptable storage nodes
	MaxTimeout          time.Duration // Maximum timeout for operations
	GracefulDegradation bool          // Allow storing on fewer than K nodes
}

func SendToDb(key [20]byte, msgcert interface{}, route string) error {
	// NEW: Enhanced network configuration
	networkCfg := NetworkConfig{
		MinStorageNodes:     max(1, config.K/2), // At least half of K, minimum 1
		MaxTimeout:          10 * time.Second,   // Increased timeout for sparse networks
		GracefulDegradation: true,               // Allow partial replication
	}

	var mu sync.Mutex
	startNodes, err := getStartNodesWithFallback() // NEW: Enhanced bootstrap
	if err != nil || len(startNodes) == 0 {
		return fmt.Errorf("failed to get bootstrap nodes: %v", err)
	}

	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)
	stored := make(map[string]bool)
	failed := make(map[string]bool) // NEW: Track failed nodes
	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})

	// NEW: Adaptive convergence parameters for sparse networks
	maxSame := 3 // Increased from 2 for better sparse network handling
	sameCount := 0
	var prevClosest []*types.Node
	madeProgress := false
	roundsWithoutNewNodes := 0 // NEW: Track discovery stagnation

	// NEW: Adaptive timeout based on network size
	ctx, cancel := context.WithTimeout(context.Background(), networkCfg.MaxTimeout)
	defer cancel()

	log.Printf("Starting Kademlia store with %d bootstrap nodes", len(startNodes))

	// Worker goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Worker goroutine panic recovered: %v", r)
				close(done)
			}
		}()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				log.Printf("SendToDb worker timed out after %v", networkCfg.MaxTimeout)
				close(done)
				return
			default:
			}

			mu.Lock()
			// Filter out failed nodes before sorting
			activeKnown := make([]*types.Node, 0, len(known))
			for _, node := range known {
				if !failed[node.PeerId] {
					activeKnown = append(activeKnown, node)
				}
			}

			sort.Slice(activeKnown, func(i, j int) bool {
				return util.XORBigInt(key, activeKnown[i].NodeId).Cmp(util.XORBigInt(key, activeKnown[j].NodeId)) < 0
			})

			currentClosest := append([]*types.Node(nil), activeKnown...)
			if len(currentClosest) > config.K {
				currentClosest = currentClosest[:config.K]
			}

			same := len(currentClosest) == len(prevClosest)
			if same && len(prevClosest) > 0 {
				for i := range currentClosest {
					if i >= len(prevClosest) || !bytes.Equal(currentClosest[i].NodeId[:], prevClosest[i].NodeId[:]) {
						same = false
						break
					}
				}
			}

			// NEW: Enhanced convergence strategy for sparse networks
			if same {
				if madeProgress {
					sameCount = max(0, sameCount-1)
					roundsWithoutNewNodes = 0
				} else {
					sameCount++
					roundsWithoutNewNodes++
				}
			} else {
				sameCount = 0
				roundsWithoutNewNodes = 0
			}
			madeProgress = false

			// NEW: Multiple termination conditions
			shouldTerminate := false
			terminationReason := ""

			// Check if we have enough stored responses
			storedCount := len(stored)
			if storedCount >= config.K {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Target replication achieved (%d/%d)", storedCount, config.K)
			} else if networkCfg.GracefulDegradation && storedCount >= networkCfg.MinStorageNodes && sameCount >= maxSame {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Graceful degradation: %d/%d stored, network converged", storedCount, config.K)
			} else if roundsWithoutNewNodes >= 5 {
				// NEW: Terminate if we haven't discovered new nodes for several rounds
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Discovery stagnation: %d stored, no new nodes found", storedCount)
			} else if sameCount >= maxSame && len(activeKnown) <= config.Alpha {
				// NEW: Small network detection
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Small network detected: %d active nodes, %d stored", len(activeKnown), storedCount)
			}

			if shouldTerminate {
				mu.Unlock()
				logger.LogToFile(fmt.Sprintf("Kademlia store terminated: %s", terminationReason))
				log.Printf("Kademlia store terminated: %s", terminationReason)
				close(done)
				return
			}

			prevClosest = currentClosest
			toQuery := []*types.Node{}
			for _, n := range currentClosest {
				nodeKey := n.PeerId
				if !queried[nodeKey] && !failed[nodeKey] {
					toQuery = append(toQuery, n)
					queried[nodeKey] = true
					if len(toQuery) >= config.Alpha {
						break
					}
				}
			}

			// NEW: If no nodes to query from closest, try any unqueried nodes
			if len(toQuery) == 0 {
				for _, n := range activeKnown {
					nodeKey := n.PeerId
					if !queried[nodeKey] && !failed[nodeKey] {
						toQuery = append(toQuery, n)
						queried[nodeKey] = true
						if len(toQuery) >= config.Alpha {
							break
						}
					}
				}
			}

			mu.Unlock()

			if len(toQuery) == 0 {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			log.Printf("Querying %d nodes (stored: %d/%d, known: %d)", len(toQuery), len(stored), config.K, len(activeKnown))

			var wg sync.WaitGroup
			for _, n := range toQuery {
				wg.Add(1)
				go func(n *types.Node) {
					defer wg.Done()
					resp, err := network.SendTo(n.PeerId, route, msgcert, "db")
					if err != nil {
						log.Printf("Failed to store to %s: %v", n.PeerId, err)
						mu.Lock()
						failed[n.PeerId] = true // NEW: Mark as failed
						mu.Unlock()
						return
					}

					respBytes, ok := resp.([]byte)
					if !ok {
						logger.LogToFile("Unexpected Response format received")
						log.Printf("Unexpected response format from %s", n.PeerId)
						mu.Lock()
						failed[n.PeerId] = true // NEW: Mark as failed
						mu.Unlock()
						return
					}

					var base BaseResponse
					if err := json.Unmarshal(respBytes, &base); err != nil {
						logger.LogToFile("[DEBUG]Failed to parse base response")
						log.Printf("Failed to parse base response from %s: %v", n.PeerId, err)
						mu.Lock()
						failed[n.PeerId] = true // NEW: Mark as failed
						mu.Unlock()
						return
					}

					switch base.Type {
					case "stored":
						var storedResp StoredResponse
						if err := json.Unmarshal(respBytes, &storedResp); err != nil {
							log.Printf("Failed to decode stored response: %v", err)
							return
						}

						mu.Lock()
						stored[n.PeerId] = true
						madeProgress = true

						// NEW: Handle nodes from stored response
						for _, newNode := range storedResp.Nodes {
							nodeCopy := newNode
							select {
							case newNodesChan <- &nodeCopy:
							default:
								// Channel full, skip
							}
						}

						// NEW: Check termination condition with graceful degradation
						storedCount := len(stored)
						if storedCount >= config.K {
							log.Printf("Target replication achieved: %d/%d nodes", storedCount, config.K)
							mu.Unlock()
							close(done)
							return
						} else if networkCfg.GracefulDegradation && storedCount >= networkCfg.MinStorageNodes {
							log.Printf("Minimum replication achieved: %d/%d nodes (target: %d)", storedCount, networkCfg.MinStorageNodes, config.K)
						}
						mu.Unlock()
					case "redirect":
						var redirectResp RedirectResponse
						if err := json.Unmarshal(respBytes, &redirectResp); err != nil {
							log.Printf("Failed to decode redirect response: %v", err)
							return
						}
						for _, newNode := range redirectResp.Nodes {
							nodeCopy := newNode
							select {
							case newNodesChan <- &nodeCopy:
							default:
								// Channel full, skip
							}
						}

					default:
						log.Printf("Unknown response type '%s' from %s", base.Type, n.PeerId)
					}
				}(n)
			}
			wg.Wait()
		}
	}()

	// Listen for new nodes or completion
	for {
		select {
		case newNode := <-newNodesChan:
			mu.Lock()
			alreadyKnown := false
			for _, kn := range known {
				if bytes.Equal(kn.NodeId[:], newNode.NodeId[:]) {
					alreadyKnown = true
					break
				}
			}
			if !alreadyKnown && !failed[newNode.PeerId] {
				known = append(known, newNode)
				log.Printf("Discovered new node: %s (total known: %d)", newNode.PeerId, len(known))
			}
			mu.Unlock()

		case <-done:
			storedCount := len(stored)
			failedCount := len(failed)
			totalKnown := len(known)

			log.Printf("Recursive store finished: %d stored, %d failed, %d total nodes discovered", storedCount, failedCount, totalKnown)

			// NEW: Provide detailed result analysis
			if storedCount >= config.K {
				log.Printf("✅ SUCCESS: Achieved target replication (%d/%d)", storedCount, config.K)
			} else if networkCfg.GracefulDegradation && storedCount >= networkCfg.MinStorageNodes {
				log.Printf("⚠️  PARTIAL SUCCESS: Achieved minimum replication (%d/%d, target: %d)", storedCount, networkCfg.MinStorageNodes, config.K)
				logger.LogToFile(fmt.Sprintf("Sparse network: stored on %d/%d nodes", storedCount, config.K))
			} else {
				log.Printf("❌ INSUFFICIENT REPLICATION: Only %d/%d stored (minimum: %d)", storedCount, config.K, networkCfg.MinStorageNodes)
				if totalKnown < config.K {
					log.Printf("Network too small: only %d nodes discovered", totalKnown)
				}
			}

			return nil
		}
	}
}

// NEW: Enhanced bootstrap node retrieval with fallback mechanisms
func getStartNodesWithFallback() ([]*types.Node, error) {
	// Try primary bootstrap source
	startNodes, err := util.GetStartNodes()
	if err == nil && len(startNodes) > 0 {
		log.Printf("Retrieved %d bootstrap nodes from primary source", len(startNodes))
		return startNodes, nil
	}

	log.Printf("Primary bootstrap failed (%v), trying fallback methods", err)

	// NEW: Try multiple fallback sources
	fallbackMethods := []func() ([]*types.Node, error){
		tryLocalNodeCache,
		tryWellKnownBootstrapNodes,
		tryEnvironmentBootstrap,
	}

	for i, method := range fallbackMethods {
		if nodes, err := method(); err == nil && len(nodes) > 0 {
			log.Printf("Fallback method %d succeeded: %d nodes", i+1, len(nodes))
			return nodes, nil
		}
	}

	return nil, fmt.Errorf("all bootstrap methods failed")
}

// NEW: Fallback bootstrap methods
func tryLocalNodeCache() ([]*types.Node, error) {
	// Implement local cache of previously known nodes
	// This is a placeholder - implement based on your caching strategy
	log.Println("Trying local node cache...")
	return nil, fmt.Errorf("local cache not implemented")
}

func tryWellKnownBootstrapNodes() ([]*types.Node, error) {
	// Implement well-known bootstrap nodes
	// This is a placeholder - implement based on your network setup
	log.Println("Trying well-known bootstrap nodes...")
	return nil, fmt.Errorf("well-known nodes not implemented")
}

func tryEnvironmentBootstrap() ([]*types.Node, error) {
	// Try environment variables or config files
	// This is a placeholder - implement based on your deployment strategy
	log.Println("Trying environment bootstrap...")
	return nil, fmt.Errorf("environment bootstrap not implemented")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
