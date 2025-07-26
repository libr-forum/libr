package bootstrap

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
)

func GetBootstrapNodes() []string {
	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/db_addresses.csv"
	bootAddrs, _ := getAllValidAddresses(csvURL)

<<<<<<< HEAD
	// var bootNodes []node.Node
	// for _, addr := range bootAddrs {
	// 	parts := strings.Split(addr, ":")
	// 	if len(parts) != 2 {
	// 		log.Printf("invalid address format: %s", addr)
	// 		continue
	// 	}
	// 	ip := parts[0]
	// 	port := parts[1]

	// 	bootNodes = append(bootNodes, node.Node{
	// 		NodeId: node.GenerateNodeID(addr),
	// 		IP:     ip,
	// 		Port:   port,
	// 		// LastSeen left as 0
	// 	})
	// }

	// fmt.Println("Boot nodes:", bootNodes)
	return bootAddrs
}

// func GetRelayAddrs() ([]string, error) {
// 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/main/docs/network.csv"

// 	resp, err := http.Get(csvURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	reader := csv.NewReader(resp.Body)
// 	var relayAddrs []string

// 	// Skip the header line
// 	_, err = reader.Read()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read CSV header: %w", err)
// 	}

// 	for {
// 		row, err := reader.Read()
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			log.Printf("skipping bad row: %v", err)
// 			continue
// 		}

// 		if len(row) < 2 {
// 			log.Printf("skipping row with too few columns: %v", row)
// 			continue
// 		}

// 		addr := row[0]
// 		relayAddrs = append(relayAddrs, addr)
// 	}

// 	if len(relayAddrs) == 0 {
// 		return nil, fmt.Errorf("no valid relay addresses found")
// 	}

// 	return relayAddrs, nil
// }

func GetRelayAddrs() ([]string, error) {
	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/network.csv"
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

	var relayAddrs []string

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("skipping bad row: %v", err)
			continue
		}

		if len(row) < 1 {
			log.Printf("skipping row with too few columns: %v", row)
			continue
		}

		relayAddrs = append(relayAddrs, row[0])
	}

	if len(relayAddrs) == 0 {
		return nil, fmt.Errorf("no valid address found")
	}
=======
	return bootAddrs
}

func GetRelayAddrs() ([]string, error) {
	// csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/network.csv"
	// resp, err := http.Get(csvURL)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to fetch CSV: %w", err)
	// }
	// defer resp.Body.Close()

	// reader := csv.NewReader(resp.Body)

	// // Skip header
	// if _, err := reader.Read(); err != nil {
	// 	return nil, fmt.Errorf("failed to read header: %w", err)
	// }

	var relayAddrs []string

	relayAddrs = append(relayAddrs, "/dns4/0.tcp.in.ngrok.io/tcp/11207/p2p/12D3KooWQTmGS67k3hoD1oL69ZsDaLBqWjMD9kGqKv8zarVZgpno")

	// for {
	// 	row, err := reader.Read()
	// 	if err != nil {
	// 		if err.Error() == "EOF" {
	// 			break
	// 		}
	// 		log.Printf("skipping bad row: %v", err)
	// 		continue
	// 	}

	// 	if len(row) < 1 {
	// 		log.Printf("skipping row with too few columns: %v", row)
	// 		continue
	// 	}

	// 	relayAddrs = append(relayAddrs, row[0])
	// }

	// if len(relayAddrs) == 0 {
	// 	return nil, fmt.Errorf("no valid address found")
	// }
>>>>>>> 33bf593 (Migrated from postgresql to sqlite)

	return relayAddrs, nil
}

func getAllValidAddresses(csvURL string) ([]string, error) {
	// Fetch the CSV file over HTTP
	resp, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	var addresses []string

	// Skip the header line
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Read and process each row
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Skipping bad row: %v", err)
			continue
		}

		if len(row) < 2 {
			log.Printf("Skipping row with too few columns: %v", row)
			continue
		}

		ip := row[0]
		port := row[1]
		address := ip + ":" + port

		// Optional: Add a health check here before accepting the address

		addresses = append(addresses, address)
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no valid addresses found")
	}

	return addresses, nil
}
