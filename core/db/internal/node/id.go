package node

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"math/big"

	"github.com/libr-forum/Libr/core/db/internal/keycache"
)

func GenerateNodeID(input string) [20]byte {
	h := sha1.New()
	h.Write([]byte(input))
	var id [20]byte
	copy(id[:], h.Sum(nil))
	return id
}

func GenerateNodeIDFromPublicKey() string {
	pubKeyB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
	nodeID := GenerateNodeID(pubKeyB64)
	nodeIDStr := base64.StdEncoding.EncodeToString(nodeID[:])
	return nodeIDStr
}

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

func DecodeNodeID(base64Str string) ([20]byte, error) {
	var id [20]byte

	bytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return id, err
	}

	if len(bytes) != 20 {
		return id, errors.New("invalid ID length: expected 20 bytes")
	}

	copy(id[:], bytes)
	return id, nil
}
