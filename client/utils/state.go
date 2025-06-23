package util

import (
	"libr/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	const publicKeyStr = "MFQKfOL+2XnO1IrZYWp0cxOV7P4DNyEOTq3dQvjgS5o="

	mods := []types.Mod{
		{
			IP:        "127.0.0.1",
			Port:      "5000",
			PublicKey: publicKeyStr,
		},
	}

	return mods, nil
}
