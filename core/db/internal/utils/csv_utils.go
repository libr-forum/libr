// package utils

// import (
// 	"encoding/csv"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"net/url"

// 	"github.com/devlup-labs/Libr/core/db/internal/models"
// )

// func GetBootstrapNodes() []string {
// 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/db_addresses.csv"
// 	bootAddrs, _ := getAllValidAddresses(csvURL)

// 	// var bootNodes []node.Node
// 	// for _, addr := range bootAddrs {
// 	// 	parts := strings.Split(addr, ":")
// 	// 	if len(parts) != 2 {
// 	// 		log.Printf("invalid address format: %s", addr)
// 	// 		continue
// 	// 	}
// 	// 	ip := parts[0]
// 	// 	port := parts[1]

// 	// 	bootNodes = append(bootNodes, node.Node{
// 	// 		NodeId: node.GenerateNodeID(addr),
// 	// 		IP:     ip,
// 	// 		Port:   port,
// 	// 		// LastSeen left as 0
// 	// 	})
// 	// }

// 	// fmt.Println("Boot nodes:", bootNodes)
// 	return bootAddrs
// }

// // func GetRelayAddrs() ([]string, error) {
// // 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/main/docs/network.csv"

// // 	resp, err := http.Get(csvURL)
// // 	if err != nil {
// // 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// // 	}
// // 	defer resp.Body.Close()

// // 	reader := csv.NewReader(resp.Body)
// // 	var relayAddrs []string

// // 	// Skip the header line
// // 	_, err = reader.Read()
// // 	if err != nil {
// // 		return nil, fmt.Errorf("failed to read CSV header: %w", err)
// // 	}

// // 	for {
// // 		row, err := reader.Read()
// // 		if err != nil {
// // 			if err.Error() == "EOF" {
// // 				break
// // 			}
// // 			log.Printf("skipping bad row: %v", err)
// // 			continue
// // 		}

// // 		if len(row) < 2 {
// // 			log.Printf("skipping row with too few columns: %v", row)
// // 			continue
// // 		}

// // 		addr := row[0]
// // 		relayAddrs = append(relayAddrs, addr)
// // 	}

// // 	if len(relayAddrs) == 0 {
// // 		return nil, fmt.Errorf("no valid relay addresses found")
// // 	}

// // 	return relayAddrs, nil
// // }

// func GetRelayAddrs() ([]string, error) {
// 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/network.csv"
// 	resp, err := http.Get(csvURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	reader := csv.NewReader(resp.Body)

// 	// Skip header
// 	if _, err := reader.Read(); err != nil {
// 		return nil, fmt.Errorf("failed to read header: %w", err)
// 	}

// 	var relayAddrs []string

// 	for {
// 		row, err := reader.Read()
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			log.Printf("skipping bad row: %v", err)
// 			continue
// 		}

// 		if len(row) < 1 {
// 			log.Printf("skipping row with too few columns: %v", row)
// 			continue
// 		}

// 		relayAddrs = append(relayAddrs, row[0])
// 	}

// 	if len(relayAddrs) == 0 {
// 		return nil, fmt.Errorf("no valid address found")
// 	}

// 	return relayAddrs, nil
// }

// func getAllValidAddresses(csvURL string) ([]string, error) {
// 	// Fetch the CSV file over HTTP
// 	resp, err := http.Get(csvURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	reader := csv.NewReader(resp.Body)
// 	var addresses []string

// 	// Skip the header line
// 	_, err = reader.Read()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read CSV header: %w", err)
// 	}

// 	// Read and process each row
// 	for {
// 		row, err := reader.Read()
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			log.Printf("Skipping bad row: %v", err)
// 			continue
// 		}

// 		if len(row) < 2 {
// 			log.Printf("Skipping row with too few columns: %v", row)
// 			continue
// 		}

// 		ip := row[0]
// 		port := row[1]
// 		address := ip + ":" + port

// 		// Optional: Add a health check here before accepting the address

// 		addresses = append(addresses, address)
// 	}

// 	if len(addresses) == 0 {
// 		return nil, fmt.Errorf("no valid addresses found")
// 	}

// 	return addresses, nil
// }

// func GetValidMods() ([]*models.Mod, error) {
// 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/mod_addresses.csv"
// 	resp, err := http.Get(csvURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	reader := csv.NewReader(resp.Body)

// 	// Skip header
// 	if _, err := reader.Read(); err != nil {
// 		return nil, fmt.Errorf("failed to read header: %w", err)
// 	}

// 	var mods []*models.Mod

// 	for {
// 		row, err := reader.Read()
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			log.Printf("skipping bad row: %v", err)
// 			continue
// 		}

// 		if len(row) < 3 {
// 			log.Printf("skipping row with too few columns: %v", row)
// 			continue
// 		}

// 		mod := models.Mod{
// 			IP:        row[0],
// 			Port:      row[1],
// 			PublicKey: row[2],
// 		}
// 		mods = append(mods, &mod)
// 	}

// 	return mods, nil
// }

// func fetchSheetData(sheetName string) ([]string, error) {
// 	endpoint := fmt.Sprintf("%s?sheet=%s", baseURL, url.QueryEscape(sheetName))
// 	resp, err := http.Get(endpoint)
// 	if err != nil {
// 		return nil, fmt.Errorf("error making request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("non-OK status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result []string
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, fmt.Errorf("error decoding JSON: %v", err)
// 	}

// 	return result, nil
// }

//	func main() {
//		sheets := []string{"relay", "mod", "db", "all mods"}
//		for _, sheet := range sheets {
//			data, err := fetchSheetData(sheet)
//			if err != nil {
//				fmt.Printf("Error fetching %s: %v\n", sheet, err)
//				continue
//			}
//			fmt.Printf("Data from '%s':\n", sheet)
//			for _, addr := range data {
//				fmt.Println(" -", addr)
//			}
//			fmt.Println()
//		}
//	}

package utils

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/devlup-labs/Libr/core/db/internal/models"
)

// const baseURL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pubhtml" // Replace with actual URL

// // Generic function to fetch raw [][]interface{} from a sheet
// func fetchRawData(sheet string) ([][]interface{}, error) {

// 	url := fmt.Sprintf("%s?sheet=%s", baseURL, sheet)
// 	fmt.Println("▶ fetching sheet:", sheet, "from URL:", url)
// 	start := time.Now()
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	elapsed := time.Since(start)
// 	fmt.Printf("⏱️ utils.GetDBAddrList() took %s\n", elapsed)

// 	defer resp.Body.Close()

// 	bodyBytes, _ := io.ReadAll(resp.Body)

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
// 	}

// 	var rows [][]interface{}
// 	if err := json.Unmarshal(bodyBytes, &rows); err != nil {
// 		return nil, fmt.Errorf("invalid JSON: %w", err)
// 	}
// 	return rows, nil
// }

// func GetSheetData(sheet string) ([][]interface{}, error) {
// 	return fetchRawData(sheet)
// }

// func GetDBAddrList() ([]string, error) {
// 	rows, err := fetchRawData("db")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var addrList []string
// 	for _, r := range rows {
// 		if len(r) >= 2 {
// 			ip := fmt.Sprint(r[0])
// 			port := fmt.Sprint(r[1])
// 			addrList = append(addrList, fmt.Sprintf("%s:%s", ip, port))
// 		}
// 	}
// 	return addrList, nil
// }

// func GetModList() ([]*models.Mod, error) {
// 	rows, err := fetchRawData("mod")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var mods []*models.Mod
// 	for _, r := range rows {
// 		if len(r) >= 3 {
// 			mod := &models.Mod{
// 				IP:        fmt.Sprint(r[0]),
// 				Port:      fmt.Sprint(r[1]),
// 				PublicKey: fmt.Sprint(r[2]),
// 			}
// 			mods = append(mods, mod)
// 		}
// 	}
// 	return mods, nil
// }

// func GetRelayAddresses() ([]string, error) {
// 	rows, err := fetchRawData("relay")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var addrs []string
// 	for _, r := range rows {
// 		if len(r) >= 1 {
// 			fmt.Println(r[0])
// 			addrs = append(addrs, fmt.Sprint(r[0]))
// 		}
// 	}
// 	return addrs, nil
// }

// func GetAllModPublicKeys() ([]string, error) {
// 	rows, err := fetchRawData("all mods")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var keys []string
// 	for _, r := range rows {
// 		if len(r) >= 1 {
// 			keys = append(keys, fmt.Sprint(r[0]))
// 		}
// 	}
// 	return keys, nil
// }

const (
	modGID   = "1379617454"
	dbGID    = "26032376"
	relayGID = "1789680527"
)

func getSheetData(gid string) ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pub?gid=%s&single=true&output=csv", gid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sheet data: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	allRows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Detect and skip header row
	if len(allRows) > 0 {
		firstRow := allRows[0]
		// Check for typical header keywords
		for _, cell := range firstRow {
			lower := strings.ToLower(cell)
			if strings.Contains(lower, "relay") || strings.Contains(lower, "mod") || strings.Contains(lower, "addr") || strings.Contains(lower, "pubkey") {
				return allRows[1:], nil // skip header
			}
		}
	}

	return allRows, nil
}

// Get mod data as []*Mod
func GetModData() ([]*models.Mod, error) {
	data, err := getSheetData(modGID)
	if err != nil {
		return nil, err
	}

	var mods []*models.Mod
	for _, row := range data {
		if len(row) >= 3 {
			mods = append(mods, &models.Mod{
				IP:        strings.TrimSpace(row[0]),
				Port:      strings.TrimSpace(row[1]),
				PublicKey: strings.TrimSpace(row[2]),
			})
		}
	}
	return mods, nil
}

// Get db data as []string (ip:port format)
func GetDbData() ([]string, error) {
	data, err := getSheetData(dbGID)
	if err != nil {
		return nil, err
	}

	var dbList []string
	for _, row := range data {
		if len(row) >= 2 {
			ip := strings.TrimSpace(row[0])
			port := strings.TrimSpace(row[1])
			addr := ip + ":" + port
			if addr != "IP:Port" && ip != "" && port != "" {
				dbList = append(dbList, addr)
			}
		}
	}
	return dbList, nil
}

// Get relay data as []string
func GetRelayData() ([]string, error) {
	data, err := getSheetData(relayGID)
	if err != nil {
		return nil, err
	}

	var relayList []string
	for _, row := range data {
		if len(row) >= 1 {
			relayList = append(relayList, strings.TrimSpace(row[0]))
		}
	}
	return relayList, nil
}
