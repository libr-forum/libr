package storage

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
)

func StoreMsgCert(msgcert *models.MsgCert) (string, error) {
	fmt.Println("Storing MsgCert")
	if err := ValidateModCert(msgcert); err != nil {
		return "", err
	}

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

	fmt.Println("Probably stored???")
	if err != nil {
		log.Printf("Error inserting MsgCert: %v", err)
		return "Error inserting MsgCert", err
	}
	return "Message certificate successfully inserted", nil
}

func DeleteMsgCert(repCert *models.ReportCert) error {
	query := "UPDATE msgcert SET deleted = 1 WHERE ts = ? AND sender = ?;"

	result, err := config.DB.Exec(query, repCert.Msgcert.Msg.Ts, repCert.Msgcert.Msg.Content)
	if err != nil {
		log.Printf("Error soft-deleting MsgCert: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("MsgCert not found")
	}

	return nil
}

func GetMsgCert(ts int64) []models.RetMsgCert {
	// Truncate the timestamp to the minute

	minute := (ts / 60) * 60
	nextMinute := minute + 60

	query := `
	SELECT sender, content, ts, mod_certs, sign, deleted
	FROM msgcert
	WHERE ts >= ? AND ts < ?
`
	rows, err := config.DB.Query(query, minute, nextMinute)
	if err != nil {
		log.Printf("Error fetching MsgCert: %v", err)
		return nil
	}
	defer rows.Close()

	var retMsgCerts []models.RetMsgCert
	for rows.Next() {
		var retMsgCert models.RetMsgCert
		var modCertsJSON string
		var tsVal int64

		if err := rows.Scan(&retMsgCert.PublicKey, &retMsgCert.Msg.Content, &tsVal, &modCertsJSON, &retMsgCert.Sign, &retMsgCert.Deleted); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		retMsgCert.Msg.Ts = tsVal

		if err := json.Unmarshal([]byte(modCertsJSON), &retMsgCert.ModCerts); err != nil {
			log.Printf("Error unmarshaling modCerts: %v", err)
			continue
		}

		retMsgCerts = append(retMsgCerts, retMsgCert)
	}

	return retMsgCerts
}
