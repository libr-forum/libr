package util

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"libr/types"
	"strconv"
)

var modPublicKeys = []string{
	"8f9a7c3efc0a1a5d4e2a7b9d1c6e4f3a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
	"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
}

func GetOnlineMods() ([]types.Mod, error) {
	mods := make([]types.Mod, 0, len(modPublicKeys))

	for i, hexKey := range modPublicKeys {
		pubKeyBytes, err := hex.DecodeString(hexKey)
		if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
			return nil, errors.New("invalid public key in modPublicKeys")
		}

		mod := types.Mod{
			IP:        "127.0.0.1",             // or dynamic discovery
			Port:      "900" + strconv.Itoa(i), // e.g., "9000", "9001"
			PublicKey: ed25519.PublicKey(pubKeyBytes),
		}

		mods = append(mods, mod)
	}

	return mods, nil
}
