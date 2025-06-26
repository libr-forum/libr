package util

import (
	"github.com/devlup-labs/Libr/core/client/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	mods := []types.Mod{
		{
			IP:        "127.0.0.1",
			Port:      "5000",
			PublicKey: "ZXLvgdRKGT467Y9QCjxyaEvG40Ryvh4nPDoHjLYYE6E=",
		},
		{
			IP:        "127.0.0.1",
			Port:      "5001",
			PublicKey: "uLSLJAx+noFAfz0mSIxonc6aD336vmSrgODwtiN1tpI=",
		},
		{
			IP:        "127.0.0.1",
			Port:      "5002",
			PublicKey: "vuBnetbCJpcHHdEYj8aZAYEvhf6Yg0PZXwjs9A5XNmA=",
		},
	}
	return mods, nil
}
