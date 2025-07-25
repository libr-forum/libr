package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func AmIMod(myKey string) (bool, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/cherry-aggarwal/LIBR/integration/docs/all_mods.csv?nocache=%d", time.Now().UnixNano())
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return false, fmt.Errorf("failed to read header: %w", err)
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("skipping bad row: %v", err)
			continue
		}

		if len(row) > 0 && row[0] == myKey {
			return true, nil
		}
	}

	return false, nil
}
