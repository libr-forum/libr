package msgcert

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/devlup-labs/Libr/core/db/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func StoreMsgCert(msgcert models.MsgCert) (string, error) {

	query := "INSERT INTO MsgCert(sender,content,ts,mod_cert,sign) VALUES ($1,$2,$3,$4,$5)"
	modCertsJSON, _ := json.Marshal(msgcert.ModCerts)

	_, err := Pool.Exec(context.Background(), query, msgcert.PublicKey, msgcert.Msg.Content, time.Unix(msgcert.Msg.Ts, 0), modCertsJSON, msgcert.Sign)
	if err != nil {
		fmt.Printf("error inserting Message certificate: %v", err)
		return "Error", err
	}
	return "Message certificate Successfully Inserted", nil
}

func GetMsgCert(ts int64) []models.MsgCert {
	query := "SELECT * FROM MsgCert WHERE ts = $1"
	rows, err := Pool.Query(context.Background(), query, ts)
	if err != nil {
		fmt.Printf("error getting MsgCert from db: %v", err)
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
	fmt.Println(msgCerts)
	return msgCerts
}
