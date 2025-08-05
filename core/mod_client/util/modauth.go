package util

import (
	"fmt"

	"github.com/devlup-labs/Libr/core/mod_client/types"
)

func AmIMod(myKey string) (bool, error) {
	rows, err := fetchRawData("1379617454")
	fmt.Println(rows)
	if err != nil {
		return false, err
	}
	var mods []types.Mod
	for _, r := range rows {
		if len(r) >= 3 {
			mod := types.Mod{
				IP:        r[0],
				Port:      r[1],
				PublicKey: r[2],
			}
			mods = append(mods, mod)
		}
	}
	for _, mod := range mods {
		if len(mod.PublicKey) > 0 && mod.PublicKey == myKey {
			return true, nil
		}
	}

	return false, nil
}
