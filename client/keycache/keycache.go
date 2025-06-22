// keycache/cache.go
package keycache

import (
	"crypto/ed25519"
	"log"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

var (
	PubKey  ed25519.PublicKey
	PrivKey ed25519.PrivateKey
)

func InitKeys() {
	var err error
	PubKey, PrivKey, err = cryptoutils.LoadKeys()
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}
}
