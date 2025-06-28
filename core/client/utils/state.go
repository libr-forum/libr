package util

import (
	"github.com/devlup-labs/Libr/core/client/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	mods := []types.Mod{
		{
			IP:        "localhost",
			Port:      "3000",
			PublicKey: "d+Mst0eUGiL2pduG3MEQXAYobGx7JG7EeSC29OsipeI=",
		},
		{
			IP:        "localhost",
			Port:      "4000",
			PublicKey: "d+Mst0eUGiL2pduG3MEQXAYobGx7JG7EeSC29OsipeI=",
		}, {
			IP:        "localhost",
			Port:      "5000",
			PublicKey: "d+Mst0eUGiL2pduG3MEQXAYobGx7JG7EeSC29OsipeI=",
		},
	}
	return mods, nil
}
