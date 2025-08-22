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

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/network"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

func Fetch(ts int64) []types.RetMsgCert {
	key := strconv.FormatInt(ts, 10)
	keyBytes := util.GenerateNodeID(key)

	startNodes, _ := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)

	var allCerts []types.RetMsgCert
	deleteCount := make(map[string]int)
	mu := sync.Mutex{}

	const maxRounds = 50
	const alpha = 3
	const k = config.K
	const deleteThreshold = 2

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
			key := n.PeerId
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

		for _, n := range toQuery {
			wg.Add(1)
			go func(n *types.Node) {
				defer wg.Done()
				rawResp, err := network.GetFrom(n.PeerId, fmt.Sprintf("/route=find_value&&ts=%d", ts), key)
				if err != nil {
					return
				}
				respBytes, ok := rawResp.([]byte)
				if !ok {
					return
				}
				fmt.Println("Received response from:", n.PeerId)
				var base BaseResponse
				if err := json.Unmarshal(respBytes, &base); err != nil {
					return
				}

				switch base.Type {
				case "found":
					var val struct {
						Type   string             `json:"type"`
						Values []types.RetMsgCert `json:"values"`
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
							ModCerts:  cert.ModCerts,
						}
						jsonBytes, _ := json.Marshal(dataToSign)

						if cryptoutils.VerifySignature(cert.PublicKey, string(jsonBytes), cert.Sign) {
							mu.Lock()
							allCerts = append(allCerts, cert)
							if cert.Deleted == "1" {
								deleteCount[cert.Sign]++
							}
							mu.Unlock()
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

	// Filter certs: keep only one per Sign, if Deleted == "0" and delete count ≤ threshold
	unique := make(map[string]types.RetMsgCert)
	for _, cert := range allCerts {
		if cert.Deleted == "0" && deleteCount[cert.Sign] <= deleteThreshold {
			if _, exists := unique[cert.Sign]; !exists {
				unique[cert.Sign] = cert
			}
		}
	}

	var results []types.RetMsgCert
	for _, cert := range unique {
		results = append(results, cert)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Msg.Ts > results[j].Msg.Ts
	})
	return results
}

func FetchRecent(ctx context.Context) []types.RetMsgCert {
	deleteThreshold := config.DeleteThreshold
	now := time.Now().Truncate(time.Minute).Unix()
	start := now - 3600

	tsChan := make(chan int64, 100)
	rawCerts := []types.RetMsgCert{}
	printed := sync.Map{}
	var mu sync.Mutex

	signCounts := make(map[string]int)
	deleteCounts := make(map[string]int)

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
					mu.Lock()
					signCounts[cert.Sign]++
					if cert.Deleted == "1" {
						deleteCounts[cert.Sign]++
					}
					if _, seen := printed.LoadOrStore(cert.Sign+"#"+fmt.Sprint(cert.Msg.Ts), true); !seen {
						rawCerts = append(rawCerts, cert)
					}
					mu.Unlock()
				}
			}
		}()
		fmt.Println(rawCerts)
	}

	wg.Wait()

	filtered := []types.RetMsgCert{}
	for _, cert := range rawCerts {
		mu.Lock()
		delCount := deleteCounts[cert.Sign]
		totalCount := signCounts[cert.Sign]
		mu.Unlock()

		if totalCount == 0 {
			continue
		}

		if float64(delCount)/float64(totalCount) <= deleteThreshold {
			filtered = append(filtered, cert)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Msg.Ts > filtered[j].Msg.Ts
	})
	fmt.Printf("[FetchRecent] collected: %d certs after filtering\n", len(filtered))
	fmt.Println(filtered)
	return filtered
}
