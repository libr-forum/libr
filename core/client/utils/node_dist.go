package util

import (
	"crypto/sha1"
	"math/big"

	"github.com/devlup-labs/Libr/core/client/types"
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

func GetStartNodes() []*types.Node {
	return []*types.Node{
		{
			NodeId: GenerateNodeID("127.0.0.1:8000"),
			IP:     "127.0.0.1",
			Port:   "8000",
		},
	}
}
