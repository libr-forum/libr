package util

import (
	"encoding/base64"
	"fmt"
	"libr/keycache"
	"libr/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	pub := keycache.PubKey
	fmt.Println("📦 Mod public key:", base64.StdEncoding.EncodeToString(pub))

	mods := []types.Mod{
		{
			IP:        "127.0.0.1",
			Port:      "5000",
			PublicKey: pub,
		},
	}

	return mods, nil
}
