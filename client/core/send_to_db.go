package core

import (
	"encoding/json"
	"libr/network"
	"libr/types"
	"log"
	"os"
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
		resDB := response.(string)
		if err != nil {
			log.Printf("Failed to send to DB node %s: %v", DBnode, err)
			continue
		}
		return resDB
	}
	return "None of the bootstrap DB online"
}
