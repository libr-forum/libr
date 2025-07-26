package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod_client/config"
	"github.com/devlup-labs/Libr/core/mod_client/network"
	"github.com/devlup-labs/Libr/core/mod_client/types"
	util "github.com/devlup-labs/Libr/core/mod_client/util"
)

func Fetch(ts int64) []types.MsgCert {
	key := strconv.FormatInt(ts, 10)
	keyBytes := util.GenerateNodeID(key)

	startNodes, _ := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)
	printed := sync.Map{}
	var results []types.MsgCert

	const maxRounds = 50
	const alpha = 3
	const k = config.K

	for round := 0; round < maxRounds; round++ {
		sort.Slice(known, func(i, j int) bool {
			return util.XORBigInt(keyBytes, known[i].NodeId).Cmp(util.XORBigInt(keyBytes, known[j].NodeId)) < 0
		})

		currentClosest := []*types.Node{}
		for _, n := range known {
			if len(currentClosest) >= k {
				break
			}
			currentClosest = append(currentClosest, n)
		}

		toQuery := []*types.Node{}
		for _, n := range currentClosest {
			key := fmt.Sprintf("%s:%s", n.IP, n.Port)
			if !queried[key] {
				toQuery = append(toQuery, n)
				queried[key] = true
				if len(toQuery) >= alpha {
					break
				}
			}
		}

		if len(toQuery) == 0 {
			break
		}

		var wg sync.WaitGroup
		newNodes := []*types.Node{}
		var mu sync.Mutex

		for _, n := range toQuery {
			wg.Add(1)
			go func(n *types.Node) {
				defer wg.Done()
				rawResp, err := network.GetFrom(n.IP, n.Port, fmt.Sprintf("/route=find_value&&ts=%d", ts), key)
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
						if cert.Sign == "" {
							continue
						}

						sort.SliceStable(cert.ModCerts, func(i, j int) bool {
							return cert.ModCerts[i].PublicKey < cert.ModCerts[j].PublicKey
						})

						dataToSign := types.DataToSign{
							Content:   cert.Msg.Content,
							Timestamp: cert.Msg.Ts,
							ModCerts:  cert.ModCerts, // sorted before signing
						}
						jsonBytes, _ := json.Marshal(dataToSign)

						if cryptoutils.VerifySignature(cert.PublicKey, string(jsonBytes), cert.Sign) {
							if _, loaded := printed.LoadOrStore(cert.Sign, true); !loaded {
								mu.Lock()
								results = append(results, cert)
								mu.Unlock()
							}
						}
					}
				case "redirect":
					var redir struct {
						Type  string       `json:"type"`
						Nodes []types.Node `json:"nodes"`
					}
					if err := json.Unmarshal(respBytes, &redir); err != nil {
						return
					}
					mu.Lock()
					for _, node := range redir.Nodes {
						exists := false
						for _, kn := range known {
							if bytes.Equal(kn.NodeId[:], node.NodeId[:]) {
								exists = true
								break
							}
						}
						if !exists {
							copy := node
							newNodes = append(newNodes, &copy)
						}
					}
					mu.Unlock()
				}
			}(n)
		}
		wg.Wait()

		if len(newNodes) == 0 {
			break
		}
		known = append(known, newNodes...)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Msg.Ts > results[j].Msg.Ts
	})
	return results
}

func FetchRecent(ctx context.Context) []types.MsgCert {
	now := time.Now().Truncate(time.Minute).Unix()
	start := now - 1200

	tsChan := make(chan int64, 100)
	results := []types.MsgCert{}
	printed := sync.Map{}
	var mu sync.Mutex

	go func() {
		for ts := now; ts >= start; ts -= 60 {
			select {
			case <-ctx.Done():
				close(tsChan)
				return
			case tsChan <- ts:
			}
		}
		close(tsChan)
	}()

	const workers = 30
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ts := range tsChan {
				certs := Fetch(ts)
				for _, cert := range certs {
					if cert.Sign == "" || cert.Msg.Ts < start || cert.Msg.Ts > now {
						continue
					}
					if _, seen := printed.LoadOrStore(cert.Sign, true); !seen {
						mu.Lock()
						results = append(results, cert)
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Msg.Ts > results[j].Msg.Ts
	})
	fmt.Printf("[FetchRecent] collected: %d certs\n", len(results))
	return results
}

func FetchRecentStreamOrdered(ctx context.Context, startTS int64) <-chan types.MsgCert {
	out := make(chan types.MsgCert, 100)
	bufferChan := make(chan types.MsgCert, 1000)

	var printed sync.Map
	tsChan := make(chan int64, 1000)

	// Feed timestamps: newest to oldest
	go func() {
		for ts := startTS; ts >= startTS-600; ts-- {
			select {
			case <-ctx.Done():
				close(tsChan)
				return
			case tsChan <- ts:
			}
		}
		close(tsChan)
	}()

	const numWorkers = 100
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ts := range tsChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				certs := Fetch(ts)
				for _, cert := range certs {
					if cert.Sign != "" {
						if _, seen := printed.LoadOrStore(cert.Sign, true); !seen {
							select {
							case bufferChan <- cert:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			}
		}()
	}

	// Reordering and flushing
	go func() {
		defer close(out)
		defer close(bufferChan)

		var buffer []types.MsgCert
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case cert, ok := <-bufferChan:
				if !ok {
					// Final flush
					sort.Slice(buffer, func(i, j int) bool {
						return buffer[i].Msg.Ts > buffer[j].Msg.Ts
					})
					for _, c := range buffer {
						select {
						case out <- c:
						case <-ctx.Done():
							return
						}
					}
					return
				}
				buffer = append(buffer, cert)

			case <-ticker.C:
				if len(buffer) > 0 {
					sort.Slice(buffer, func(i, j int) bool {
						return buffer[i].Msg.Ts > buffer[j].Msg.Ts
					})
					toEmit := buffer
					if len(buffer) > 10 {
						toEmit = buffer[:10]
						buffer = buffer[10:]
					} else {
						buffer = nil
					}
					for _, c := range toEmit {
						select {
						case out <- c:
						case <-ctx.Done():
							return
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		// bufferChan will be closed by the flusher above
	}()

	return out
}
