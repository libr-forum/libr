package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/devlup-labs/Libr/core/client/config"
	"github.com/devlup-labs/Libr/core/client/network"
	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/utils"
)

func Fetch(ts int64) {
	key := strconv.FormatInt(ts, 10)
	keyBytes := util.GenerateNodeID(key)

	startNodes := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)

	var mu sync.Mutex
	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})
	var printed sync.Map // deduplication by sign

	const maxSame = 2
	sameCount := 0
	var prevClosest []*types.Node

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			mu.Lock()
			// Sort by XOR distance
			sort.Slice(known, func(i, j int) bool {
				return util.XORBigInt(keyBytes, known[i].NodeId).Cmp(util.XORBigInt(keyBytes, known[j].NodeId)) < 0
			})

			// Check convergence
			currentClosest := append([]*types.Node(nil), known...)
			if len(currentClosest) > config.K {
				currentClosest = currentClosest[:config.K]
			}

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

			var once sync.Once
			var wg sync.WaitGroup
			for _, n := range toQuery {
				wg.Add(1)
				go func(n *types.Node) {
					defer wg.Done()
					rawResp, err := network.GetFrom(n.IP, n.Port, "find_value", key)
					if err != nil {
						log.Printf("Error contacting %s:%s: %v", n.IP, n.Port, err)
						return
					}

					respBytes, ok := rawResp.([]byte)
					if !ok {
						log.Printf("Invalid response format from %s:%s", n.IP, n.Port)
						return
					}

					var base BaseResponse
					if err := json.Unmarshal(respBytes, &base); err != nil {
						log.Printf("Failed to decode base response from %s:%s", n.IP, n.Port)
						return
					}

					switch base.Type {
					case "found":
						var val struct {
							Type   string          `json:"type"`
							Values []types.MsgCert `json:"values"`
						}
						if err := json.Unmarshal(respBytes, &val); err != nil {
							log.Printf("Failed to decode MsgCert from %s:%s", n.IP, n.Port)
							return
						}

						for _, cert := range val.Values {
							if _, alreadyPrinted := printed.LoadOrStore(cert.Sign, true); !alreadyPrinted {
								fmt.Printf("\nSender: %s\n%s\nTime: %d\n", cert.PublicKey, cert.Msg.Content, cert.Msg.Ts)
							}
						}

						once.Do(func() {
							close(done)
						})

					case "redirect":
						var redir struct {
							Type  string       `json:"type"`
							Nodes []types.Node `json:"nodes"`
						}
						if err := json.Unmarshal(respBytes, &redir); err != nil {
							log.Printf("Failed to decode redirect nodes from %s:%s", n.IP, n.Port)
							return
						}
						for _, nn := range redir.Nodes {
							copy := nn
							newNodesChan <- &copy
						}

					default:
						log.Printf("Unknown response type from %s:%s: %s", n.IP, n.Port, base.Type)
					}
				}(n)
			}
			wg.Wait()
		}
	}()

	// New node receiver
	for {
		select {
		case <-done:
			return
		case newNode := <-newNodesChan:
			mu.Lock()
			exists := false
			for _, kn := range known {
				if bytes.Equal(kn.NodeId[:], newNode.NodeId[:]) {
					exists = true
					break
				}
			}
			if !exists {
				known = append(known, newNode)
			}
			mu.Unlock()
		}
	}
}

func FetchRecent(limit int) {
	now := time.Now().Unix()
	oneHourAgo := now - 3600

	var collected sync.Map
	var mu sync.Mutex
	var results []types.MsgCert

	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})
	known := util.GetStartNodes()
	queried := make(map[string]bool)

	keyBytes := util.GenerateNodeID(strconv.FormatInt(now, 10))
	const maxSame = 2
	sameCount := 0
	var prevClosest []*types.Node

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			mu.Lock()
			sort.Slice(known, func(i, j int) bool {
				return util.XORBigInt(keyBytes, known[i].NodeId).Cmp(util.XORBigInt(keyBytes, known[j].NodeId)) < 0
			})

			currentClosest := append([]*types.Node(nil), known...)
			if len(currentClosest) > config.K {
				currentClosest = currentClosest[:config.K]
			}

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
				close(done)
				return
			}

			prevClosest = currentClosest
			toQuery := []*types.Node{}
			for _, n := range currentClosest {
				addr := fmt.Sprintf("%s:%s", n.IP, n.Port)
				if !queried[addr] {
					queried[addr] = true
					toQuery = append(toQuery, n)
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
					for ts := now; ts >= oneHourAgo; ts-- {
						rawResp, err := network.GetFrom(n.IP, n.Port, "find_value", strconv.FormatInt(ts, 10))
						if err != nil {
							continue
						}

						respBytes, ok := rawResp.([]byte)
						if !ok {
							continue
						}

						var base BaseResponse
						if err := json.Unmarshal(respBytes, &base); err != nil {
							continue
						}

						if base.Type == "found" {
							var val struct {
								Type   string          `json:"type"`
								Values []types.MsgCert `json:"values"`
							}
							if err := json.Unmarshal(respBytes, &val); err != nil {
								continue
							}
							for _, cert := range val.Values {
								if cert.Msg.Ts >= oneHourAgo && cert.Msg.Ts <= now {
									if _, loaded := collected.LoadOrStore(cert.Sign, true); !loaded {
										mu.Lock()
										results = append(results, cert)
										mu.Unlock()
									}
								}
							}
						} else if base.Type == "redirect" {
							var redir struct {
								Type  string       `json:"type"`
								Nodes []types.Node `json:"nodes"`
							}
							if err := json.Unmarshal(respBytes, &redir); err == nil {
								for _, nn := range redir.Nodes {
									copy := nn
									newNodesChan <- &copy
								}
							}
						}
					}
				}(n)
			}
			wg.Wait()
		}
	}()

	// Collect new nodes
	for {
		select {
		case <-done:
			mu.Lock()
			// Sort by time descending
			sort.Slice(results, func(i, j int) bool {
				return results[i].Msg.Ts > results[j].Msg.Ts
			})
			// Limit
			if len(results) > limit {
				results = results[:limit]
			}
			// Print
			for _, cert := range results {
				fmt.Printf("\nSender: %s\n%s\nTime: %d\n", cert.PublicKey, cert.Msg.Content, cert.Msg.Ts)
			}
			mu.Unlock()
			return

		case newNode := <-newNodesChan:
			mu.Lock()
			already := false
			for _, kn := range known {
				if bytes.Equal(kn.NodeId[:], newNode.NodeId[:]) {
					already = true
					break
				}
			}
			if !already {
				known = append(known, newNode)
			}
			mu.Unlock()
		}
	}
}
