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
	cache "github.com/libr-forum/Libr/core/mod_client/core/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/network"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

// Fetch retrieves all message certs stored at the given timestamp key using a specific route.
func Fetch(ts int64, route string) []types.RetMsgCert {
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
			if !queried[n.PeerId] {
				toQuery = append(toQuery, n)
				queried[n.PeerId] = true
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
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[RECOVER] Fetch worker panic: %v\n", r)
					}
				}()
				rawResp, err := network.GetFrom(n.PeerId, fmt.Sprintf("%s&&ts=%d", route, ts), key)
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

// FetchReports retrieves all report certs stored at the given timestamp key.
func FetchReports(ts int64) []types.ReportCert {
	key := strconv.FormatInt(ts, 10)
	keyBytes := util.GenerateNodeID(key)

	startNodes, _ := util.GetStartNodes()
	known := append([]*types.Node{}, startNodes...)
	queried := make(map[string]bool)

	var allCerts []types.ReportCert
	mu := sync.Mutex{}

	const maxRounds = 30
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
			if !queried[n.PeerId] {
				toQuery = append(toQuery, n)
				queried[n.PeerId] = true
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
				rawResp, err := network.GetFrom(n.PeerId, fmt.Sprintf("/route=find_report&&ts=%d", ts), key)
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
						Type   string             `json:"type"`
						Values []types.ReportCert `json:"values"`
					}
					if err := json.Unmarshal(respBytes, &val); err != nil {
						return
					}

					mu.Lock()
					allCerts = append(allCerts, val.Values...)
					mu.Unlock()

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

	unique := make(map[string]types.ReportCert)
	for _, cert := range allCerts {
		if _, exists := unique[cert.Msgcert.Sign]; !exists {
			unique[cert.Msgcert.Sign] = cert
		}
	}

	var results []types.ReportCert
	for _, cert := range unique {
		results = append(results, cert)
	}
	return results
}

// FetchRecent retrieves all message certs from the last hour.
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
		defer func() { recover() }()
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

	const workers = 10 // fewer workers on mobile to save battery
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ts := range tsChan {
				certs := Fetch(ts, "/route=find_value")
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

	// Merge with recently sent messages that might not have propagated through DHT yet
	sentCerts := cache.GetRecentSentMessages()
	seenSigns := make(map[string]struct{})
	for _, m := range filtered {
		seenSigns[m.Sign] = struct{}{}
	}

	for _, sc := range sentCerts {
		if _, seen := seenSigns[sc.Sign]; !seen {
			filtered = append(filtered, types.RetMsgCert{
				PublicKey: sc.PublicKey,
				Msg:       sc.Msg,
				ModCerts:  sc.ModCerts,
				Sign:      sc.Sign,
				Deleted:   "",
			})
		}
	}

	// Sort again after merging
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Msg.Ts > filtered[j].Msg.Ts
	})

	return filtered
}

// FetchRecentReports retrieves all report certs from the last 2 hours.
func FetchRecentReports(ctx context.Context) []types.ReportCert {
	now := time.Now().Truncate(time.Minute).Unix()
	start := now - 7200

	tsChan := make(chan int64, 100)
	var allReports []types.ReportCert
	var mu sync.Mutex

	go func() {
		defer func() { recover() }()
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

	const workers = 5
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ts := range tsChan {
				reports := FetchReports(ts)
				mu.Lock()
				allReports = append(allReports, reports...)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	unique := make(map[string]types.ReportCert)
	for _, r := range allReports {
		unique[r.Msgcert.Sign] = r
	}

	filtered := []types.ReportCert{}
	for _, r := range unique {
		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Msgcert.Msg.Ts > filtered[j].Msgcert.Msg.Ts
	})
	return filtered
}
