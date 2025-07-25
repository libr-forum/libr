package keycache

import (
	"crypto/ed25519"
	"encoding/base64"
	"log"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
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

func LoadPubKey() string {
	pub, _, _ := cryptoutils.LoadKeys()
	return base64.StdEncoding.EncodeToString(pub)
}
