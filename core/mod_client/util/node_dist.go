package util

import (
	"crypto/sha1"
	"encoding/csv"
	"fmt"
	"log"
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
	csvurl := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/db_addresses.csv"
	nodes, err := getValidDBs(csvurl)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func getValidDBs(csvURL string) ([]*types.Node, error) {
	resp, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var nodes []*types.Node

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("skipping bad row: %v", err)
			continue
		}

		if len(row) < 2 {
			log.Printf("skipping row with too few columns: %v", row)
			continue
		}

		node := &types.Node{
			NodeId: GenerateNodeID(row[0] + row[1]),
			IP:     row[0],
			Port:   row[1],
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid address found")
	}

	return nodes, nil
}
