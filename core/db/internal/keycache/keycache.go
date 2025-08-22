package keycache

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
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
	fmt.Println("private:", base64.StdEncoding.EncodeToString(PrivKey))
	fmt.Println("public:", base64.StdEncoding.EncodeToString(PubKey))
}

func LoadPubKey() string {
	pub, _, _ := cryptoutils.LoadKeys()
	return base64.StdEncoding.EncodeToString(pub)
}
