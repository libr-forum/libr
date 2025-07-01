package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func SaveRoutingTableToDB(rt *routing.RoutingTable) {
	go func() {
		data, err := json.Marshal(rt)
		if err != nil {
			fmt.Println("Failed to marshal routing table:", err)
			return
		}

		_, err = config.Pool.Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS routing_table (
				node_id TEXT PRIMARY KEY,
				data JSONB NOT NULL
			)
		`)
		if err != nil {
			fmt.Println("Error creating routing_table:", err)
			return
		}

		_, err = config.Pool.Exec(context.Background(), `
			INSERT INTO routing_table (node_id, data)
			VALUES ($1, $2)
			ON CONFLICT (node_id) DO UPDATE SET data = EXCLUDED.data
		`, rt.SelfIDHex(), string(data))

		if err != nil {
			fmt.Println("Error saving routing table:", err)
		}
	}()
}

func LoadRoutingTableFromDB(nodeID string) (*routing.RoutingTable, error) {
	var data string
	err := config.Pool.QueryRow(context.Background(), `
		SELECT data FROM routing_table WHERE node_id = $1
	`, nodeID).Scan(&data)

	if err != nil {
		return nil, errors.New("no routing table found in db")
	}

	var rt routing.RoutingTable
	if err := json.Unmarshal([]byte(data), &rt); err != nil {
		return nil, errors.New("error unmarshalling routing table from db")
	}
	return &rt, nil
}
