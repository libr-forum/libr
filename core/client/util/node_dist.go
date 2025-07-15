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
			NodeId: GenerateNodeID("49.36.179.166:53643"),
			IP:     "49.36.179.166",
			Port:   "53643",
		},
		{
			NodeId: GenerateNodeID("49.36.179.166:34665"),
			IP:     "49.36.179.166",
			Port:   "34665",
		},
	}
}
