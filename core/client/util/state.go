package util

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/client/types"
)

func GetOnlineMods() ([]types.Mod, error) {
	csvurl := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/mod_addresses.csv"
	mods, err := getValidMods(csvurl)
	if err != nil {
		return nil, err
	}
	return mods, nil
}

func getValidMods(csvURL string) ([]types.Mod, error) {
	resp, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var mods []types.Mod

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("skipping bad row: %v", err)
			continue
		}

		if len(row) < 3 {
			log.Printf("skipping row with too few columns: %v", row)
			continue
		}

		mod := types.Mod{
			IP:        row[0],
			Port:      row[1],
			PublicKey: row[2],
		}
		mods = append(mods, mod)
	}

	if len(mods) == 0 {
		return nil, fmt.Errorf("no valid address found")
	}

	return mods, nil
}
