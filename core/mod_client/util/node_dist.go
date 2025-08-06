package util

import (
	"crypto/sha1"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"net/http"
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

func fetchRawData(gid string) ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pub?output=csv&gid=%s", gid)
	fmt.Println("▶ fetching sheet:", gid, "from URL:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no data rows in sheet")
	}

	return records[1:], nil // 👈 skip the header row
}
