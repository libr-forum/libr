package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
)

func StoreMsgCert(msgcert models.MsgCert) (string, error) {
	query := "INSERT INTO msgcert(sender,content,ts,mod_certs,sign) VALUES ($1,$2,$3,$4,$5)"

	modCertsJSON, err := json.Marshal(msgcert.ModCerts)
	if err != nil {
		return "Error marshaling modCerts", err
	}
	fmt.Println(msgcert)

	_, err = config.Pool.Exec(context.Background(), query,
		msgcert.PublicKey,
		msgcert.Msg.Content,
		time.Unix(msgcert.Msg.Ts, 0),
		modCertsJSON,
		msgcert.Sign,
	)
	if err != nil {
		log.Printf("Error inserting MsgCert: %v", err)
		return "Error inserting MsgCert", err
	}
	return "Message certificate successfully inserted", nil
}

func GetMsgCert(ts int64) []models.MsgCert {
	// Truncate the given timestamp to the minute
	truncatedTime := time.Unix(ts, 0).Truncate(time.Minute)

	// PostgreSQL query to match timestamps up to the minute
	query := `SELECT sender, content, ts, mod_certs, sign 
	          FROM msgcert 
	          WHERE date_trunc('minute', ts) = $1`

	rows, err := config.Pool.Query(context.Background(), query, truncatedTime)
	if err != nil {
		log.Printf("Error fetching MsgCert: %v", err)
		return nil
	}
	defer rows.Close()

	var msgCerts []models.MsgCert
	for rows.Next() {
		var msgCert models.MsgCert
		var content string
		var dbTime time.Time
		var modCertsJSON []byte

		if err := rows.Scan(&msgCert.PublicKey, &content, &dbTime, &modCertsJSON, &msgCert.Sign); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		msgCert.Msg.Content = content
		msgCert.Msg.Ts = dbTime.Unix()

		if err := json.Unmarshal(modCertsJSON, &msgCert.ModCerts); err != nil {
			log.Printf("Error unmarshaling modCerts: %v", err)
			continue
		}

		msgCerts = append(msgCerts, msgCert)
	}

	return msgCerts
}
