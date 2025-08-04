package util

import (
	"fmt"

	"github.com/devlup-labs/Libr/core/mod_client/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	rows, err := fetchRawData("mod")
	if err != nil {
		return nil, err
	}
	var mods []types.Mod
	for _, r := range rows {
		if len(r) >= 3 {
			mod := types.Mod{
				IP:        fmt.Sprint(r[0]),
				Port:      fmt.Sprint(r[1]),
				PublicKey: fmt.Sprint(r[2]),
			}
			mods = append(mods, mod)
		}
	}
	return mods, nil
}
