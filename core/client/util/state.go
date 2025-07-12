package util

import (
	"github.com/devlup-labs/Libr/core/client/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	mods := []types.Mod{
		{
			IP:        "localhost",
			Port:      "3000",
			PublicKey: "Jl6u0CVdfVDfP9I56praRtqwn6uUuo4K3Wnt69aOwWo=",
		},
		{
			IP:        "localhost",
			Port:      "4000",
			PublicKey: "Jl6u0CVdfVDfP9I56praRtqwn6uUuo4K3Wnt69aOwWo=",
		}, {
			IP:        "localhost",
			Port:      "5000",
			PublicKey: "Jl6u0CVdfVDfP9I56praRtqwn6uUuo4K3Wnt69aOwWo=",
		},
	}
	return mods, nil
}
