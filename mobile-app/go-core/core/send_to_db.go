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
	Nodes  []types.Node `json:"nodes,omitempty"`
}

func SendToDb(key [20]byte, msgcert interface{}, route string) error {
	minStorageNodes := max(1, config.K/2)
	maxTimeout := 10 * time.Second

	var mu sync.Mutex
	startNodes, err := util.GetStartNodes()
	if err != nil || len(startNodes) == 0 {
		return fmt.Errorf("failed to get bootstrap nodes: %v", err)
	}

	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)
	stored := make(map[string]bool)
	failed := make(map[string]bool)
	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})

	maxSame := 3
	sameCount := 0
	var prevClosest []*types.Node
	madeProgress := false
	roundsWithoutNewNodes := 0

	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout)
	defer cancel()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Worker goroutine panic: %v", r)
				close(done)
			}
		}()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				log.Printf("SendToDb timed out")
				close(done)
				return
			default:
			}

			mu.Lock()
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

			storedCount := len(stored)
			shouldTerminate := false
			var terminationReason string

			if storedCount >= config.K {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Target replication achieved (%d/%d)", storedCount, config.K)
			} else if storedCount >= minStorageNodes && sameCount >= maxSame {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Graceful degradation: %d stored, converged", storedCount)
			} else if roundsWithoutNewNodes >= 5 {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Discovery stagnation: %d stored", storedCount)
			} else if sameCount >= maxSame && len(activeKnown) <= config.Alpha {
				shouldTerminate = true
				terminationReason = fmt.Sprintf("Small network: %d nodes, %d stored", len(activeKnown), storedCount)
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

			var wg sync.WaitGroup
			for _, n := range toQuery {
				wg.Add(1)
				go func(n *types.Node) {
					defer wg.Done()
					resp, err := network.SendTo(n.PeerId, route, msgcert, "db")
					if err != nil {
						mu.Lock()
						failed[n.PeerId] = true
						mu.Unlock()
						return
					}

					respBytes, ok := resp.([]byte)
					if !ok {
						mu.Lock()
						failed[n.PeerId] = true
						mu.Unlock()
						return
					}

					var base BaseResponse
					if err := json.Unmarshal(respBytes, &base); err != nil {
						mu.Lock()
						failed[n.PeerId] = true
						mu.Unlock()
						return
					}

					switch base.Type {
					case "stored":
						var storedResp StoredResponse
						if err := json.Unmarshal(respBytes, &storedResp); err != nil {
							return
						}
						mu.Lock()
						stored[n.PeerId] = true
						madeProgress = true
						for _, newNode := range storedResp.Nodes {
							nodeCopy := newNode
							select {
							case newNodesChan <- &nodeCopy:
							default:
							}
						}
						if len(stored) >= config.K {
							mu.Unlock()
							close(done)
							return
						}
						mu.Unlock()

					case "redirect":
						var redirectResp RedirectResponse
						if err := json.Unmarshal(respBytes, &redirectResp); err != nil {
							return
						}
						for _, newNode := range redirectResp.Nodes {
							nodeCopy := newNode
							select {
							case newNodesChan <- &nodeCopy:
							default:
							}
						}
					}
				}(n)
			}
			wg.Wait()
		}
	}()

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
			}
			mu.Unlock()

		case <-done:
			storedCount := len(stored)
			log.Printf("SendToDb finished: %d stored", storedCount)
			return nil
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
