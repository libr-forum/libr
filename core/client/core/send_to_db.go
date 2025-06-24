package core

import (
	"encoding/json"
	"log"
	"os"

	"github.com/devlup-labs/Libr/core/client/network"
	"github.com/devlup-labs/Libr/core/client/types"
)

func SendToDb(msgcert types.MsgCert) string {
	var bootDB []string

	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	if err := json.Unmarshal(data, &bootDB); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	for _, DBnode := range bootDB {
		response, err := network.SendTo(DBnode, msgcert, "db")
		if err != nil {
			log.Printf("Failed to send to DB node %s: %v", DBnode, err)
			continue
		}
		resDB, ok := response.(string)
		if ok {
			return resDB
		}
	}

	return "❌ None of the bootstrap DBs responded"
}
