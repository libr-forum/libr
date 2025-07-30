package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/devlup-labs/Libr/core/mod_client/config"
	"github.com/devlup-labs/Libr/core/mod_client/network"
	"github.com/devlup-labs/Libr/core/mod_client/types"
	util "github.com/devlup-labs/Libr/core/mod_client/util"
)

type BaseResponse struct {
	Type string `json:"type"`
}

type RedirectResponse struct {
	Type  string       `json:"type"`
	Nodes []types.Node `json:"nodes"`
}

type StoredResponse struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

func SendToDb(key [20]byte, msgcert interface{}, route string) error {
	var mu sync.Mutex
	startNodes, _ := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)
	stored := make(map[string]bool)
	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})

	const maxSame = 2
	sameCount := 0
	var prevClosest []*types.Node

	// Worker goroutine
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			mu.Lock()
			sort.Slice(known, func(i, j int) bool {
				return util.XORBigInt(key, known[i].NodeId).Cmp(util.XORBigInt(key, known[j].NodeId)) < 0
			})

			currentClosest := append([]*types.Node(nil), known...)
			if len(currentClosest) > config.K {
				currentClosest = currentClosest[:config.K]
			}

			// Convergence check
			same := len(currentClosest) == len(prevClosest)
			if same {
				for i := range currentClosest {
					if !bytes.Equal(currentClosest[i].NodeId[:], prevClosest[i].NodeId[:]) {
						same = false
						break
					}
				}
			}

			if same {
				sameCount++
			} else {
				sameCount = 0
			}

			if sameCount >= maxSame {
				mu.Unlock()
				log.Println("Kademlia store converged. Terminating search.")
				close(done)
				return
			}

			prevClosest = currentClosest
			toQuery := []*types.Node{}
			for _, n := range currentClosest {
				key := fmt.Sprintf("%s:%s", n.IP, n.Port)
				if !queried[key] {
					toQuery = append(toQuery, n)
					queried[key] = true
					if len(toQuery) >= config.Alpha {
						break
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
					resp, err := network.SendTo(n.IP, n.Port, route, msgcert, "db")
					if err != nil {
						log.Printf("Failed to store to %s:%s: %v", n.IP, n.Port, err)
						return
					}

					respBytes, ok := resp.([]byte)
					if !ok {
						log.Printf("Unexpected response format from %s:%s", n.IP, n.Port)
						log.Println(ok)
						return
					}

					var base BaseResponse
					if err := json.Unmarshal(respBytes, &base); err != nil {
						log.Printf("Failed to parse base response from %s:%s: %v", n.IP, n.Port, err)
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
						stored[fmt.Sprintf("%s:%s", n.IP, n.Port)] = true
						if len(stored) >= config.K {
							mu.Unlock()
							close(done)
							return
						}
						mu.Unlock()

					case "redirect":
						var redirectResp RedirectResponse
						if err := json.Unmarshal(respBytes, &redirectResp); err != nil {
							log.Printf("Failed to decode redirect response: %v", err)
							return
						}
						for _, newNode := range redirectResp.Nodes {
							copy := newNode
							newNodesChan <- &copy
						}

					default:
						log.Printf("Unknown response type '%s' from %s:%s", base.Type, n.IP, n.Port)
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
			if !alreadyKnown {
				known = append(known, newNode)
			}
			mu.Unlock()

		case <-done:
			log.Printf("Recursive store finished with %d stored responses.", len(stored))
			return nil
		}
	}
}
