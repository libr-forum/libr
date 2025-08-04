package util

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/devlup-labs/Libr/core/mod_client/types"
)

func XOR(a, b [20]byte) [20]byte {
	var result [20]byte
	for i := 0; i < 20; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func XORBigInt(a, b [20]byte) *big.Int {
	xor := XOR(a, b)
	return new(big.Int).SetBytes(xor[:])
}

func GenerateNodeID(input string) [20]byte {
	h := sha1.New()
	h.Write([]byte(input))
	var id [20]byte
	copy(id[:], h.Sum(nil))
	return id
}

func GetStartNodes() ([]*types.Node, error) {
	rows, err := fetchRawData("db")
	if err != nil {
		return nil, err
	}
	var nodeList []*types.Node
	for _, r := range rows {
		if len(r) >= 2 {
			ip := fmt.Sprint(r[0])
			port := fmt.Sprint(r[1])
			addr := fmt.Sprintf("%s:%s", ip, port)

			nodeList = append(nodeList, &types.Node{
				NodeId: GenerateNodeID(addr),
				IP:     ip,
				Port:   port,
			})
		}
	}

	return nodeList, nil
}

func fetchRawData(sheet string) ([][]interface{}, error) {
	url := fmt.Sprintf("%s?sheet=%s", "https://script.google.com/macros/s/AKfycbw5yRBiPoDTWsqMcQLhEeaxRnW2UJwscjuNNLKH5juziwAdwPrsvUh7Uzci-UhTSpOzKg/exec", sheet)
	fmt.Println("▶ fetching sheet:", sheet, "from URL:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rows [][]interface{}
	if err := json.Unmarshal(bodyBytes, &rows); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return rows, nil
}
