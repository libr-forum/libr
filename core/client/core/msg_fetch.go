package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/devlup-labs/Libr/core/client/config"
	"github.com/devlup-labs/Libr/core/client/network"
	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"
)

func Fetch(ts int64) []types.MsgCert {
	key := strconv.FormatInt(ts, 10)
	keyBytes := util.GenerateNodeID(key)

	startNodes := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)

	var mu sync.Mutex
	newNodesChan := make(chan *types.Node, 100)
	done := make(chan struct{})
	var printed sync.Map // deduplication by sign

	var results []types.MsgCert
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
						return
					}

					respBytes, ok := rawResp.([]byte)
					if !ok {
						return
					}

					var base BaseResponse
					if err := json.Unmarshal(respBytes, &base); err != nil {
						return
					}

					switch base.Type {
					case "found":
						var val struct {
							Type   string          `json:"type"`
							Values []types.MsgCert `json:"values"`
						}
						if err := json.Unmarshal(respBytes, &val); err != nil {
							return
						}

						for _, cert := range val.Values {
							if _, alreadyPrinted := printed.LoadOrStore(cert.Sign, true); !alreadyPrinted {
								mu.Lock()
								results = append(results, cert)
								mu.Unlock()
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
							return
						}
						for _, nn := range redir.Nodes {
							copy := nn
							newNodesChan <- &copy
						}
					}
				}(n)
			}
			wg.Wait()
		}
	}()

	for {
		select {
		case <-done:
			return results
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

func FetchRecent() []types.MsgCert {
	const interval = int64(1)       // seconds
	const duration = int64(10 * 60) // last 10 minutes

	end := time.Now().Unix()
	start := end - duration

	var wg sync.WaitGroup
	resultsChan := make(chan []types.MsgCert, duration/interval+1)

	for ts := start; ts <= end; ts += interval {
		wg.Add(1)
		go func(timestamp int64) {
			defer wg.Done()
			resultsChan <- Fetch(timestamp)
			fmt.Println("Trying fetch for timestamp:", timestamp)
		}(ts)

	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var finalResults []types.MsgCert
	for batch := range resultsChan {
		finalResults = append(finalResults, batch...)
	}

	return finalResults
}
