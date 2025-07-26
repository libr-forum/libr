package storage

import (
	"encoding/json"
	"log"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
)

func StoreMsgCert(msgcert models.MsgCert) (string, error) {
	query := "INSERT INTO msgcert(sender, content, ts, mod_certs, sign) VALUES (?, ?, ?, ?, ?)"

	modCertsJSON, err := json.Marshal(msgcert.ModCerts)
	if err != nil {
		return "Error marshaling modCerts", err
	}

	_, err = config.DB.Exec(query,
		msgcert.PublicKey,
		msgcert.Msg.Content,
		msgcert.Msg.Ts,
		string(modCertsJSON),
		msgcert.Sign,
	)
	if err != nil {
		log.Printf("Error inserting MsgCert: %v", err)
		return "Error inserting MsgCert", err
	}
	return "Message certificate successfully inserted", nil
}

func DeleteMsgCert(msgcert models.MsgCert) (string, error) {
	query := "UPDATE msgcert SET deleted = 1 WHERE sign = ?"

	result, err := config.DB.Exec(query, msgcert.Sign)
	if err != nil {
		log.Printf("Error soft-deleting MsgCert: %v", err)
		return "Error soft-deleting MsgCert", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching rows affected: %v", err)
		return "Error fetching delete result", err
	}

	if rowsAffected == 0 {
		return "No message certificate found with that signature", nil
	}

	return "Message certificate successfully marked as deleted", nil
}

func GetMsgCert(ts int64) []models.MsgCert {
	// Truncate the timestamp to the minute
	truncatedTs := (ts / 60) * 60

	query := `
		SELECT sender, content, ts, mod_certs, sign 
		FROM msgcert 
		WHERE (ts / 60) * 60 = ?
	`

	rows, err := config.DB.Query(query, truncatedTs)
	if err != nil {
		log.Printf("Error fetching MsgCert: %v", err)
		return nil
	}
	defer rows.Close()

	var msgCerts []models.MsgCert
	for rows.Next() {
		var msgCert models.MsgCert
		var modCertsJSON string
		var tsVal int64

		if err := rows.Scan(&msgCert.PublicKey, &msgCert.Msg.Content, &tsVal, &modCertsJSON, &msgCert.Sign); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		msgCert.Msg.Ts = tsVal

		if err := json.Unmarshal([]byte(modCertsJSON), &msgCert.ModCerts); err != nil {
			log.Printf("Error unmarshaling modCerts: %v", err)
			continue
		}

		msgCerts = append(msgCerts, msgCert)
	}

	return msgCerts
}
