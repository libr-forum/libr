package moddb

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/keycache"
	"github.com/libr-forum/Libr/core/mod_client/models"
	"github.com/libr-forum/Libr/core/mod_client/types"
)

// func StoreMsgResult(cert models.MsgCert) ([]byte, string, error) {
// 	insertQuery := `
// 	INSERT INTO msgresult (sign, content, reason)
// 	VALUES (?, ?, ?);`

// 	_, err := config.DB.Exec(insertQuery, cert.Sign, cert.Msg.Content, cert.Reason)
// 	if err == nil {
// 		return nil, "acknowledged", nil
// 	}

// 	var moderated int
// 	var modsign string
// 	var sign string

// 	row := config.DB.QueryRow(`
// 		SELECT sign, moderated, modsign
// 		FROM msgresult
// 		WHERE sign = ?;`, cert.Sign)

// 	err = row.Scan(&sign, &moderated, &modsign)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, "", fmt.Errorf("no record found after insert failed: %w", err)
// 		}
// 		return nil, "", fmt.Errorf("failed to scan existing record: %w", err)
// 	}

// 	if moderated == 1 && modsign != "" {
// 		payload := fmt.Sprintf("%d", moderated) + modsign

// 		// Load keys to sign
// 		pub, priv, err := cryptoutils.LoadKeys()
// 		if err != nil {
// 			log.Printf("Key load error: %v", err)
// 			return nil, "", fmt.Errorf("failed to load keys: %w", err)
// 		}

// 		_, signed, err := cryptoutils.SignMessage(priv, payload)
// 		if err != nil {
// 			return nil, "", fmt.Errorf("signing sign-moderated: %w", err)
// 		}

// 		return pub, signed, nil
// 	}

// 	// Otherwise, return acknowledged
// 	return nil, "acknowledged", nil
// }

func StoreMsgResult(cert types.MsgCert) (*models.ModResponse, error) {
	fmt.Println("Trying to store message result:")
	insertQuery := `
    INSERT INTO msgresult (sign, content, reason, ts, type)
    VALUES (?, ?, ?, ?, ?);`

	_, err := config.DB.Exec(insertQuery, cert.Sign, cert.Msg.Content, cert.Reason, cert.Msg.Ts, cert.Type)
	if err == nil {
		fmt.Println("Message result stored successfully")
		return &models.ModResponse{
			Sign:      cert.Sign,
			PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
			Status:    "acknowledged",
		}, nil
	}
	log.Printf("Insert failed: %v", err)
	var moderated sql.NullInt64
	var modsign sql.NullString
	var sign string
	var ts int64
	var existingType sql.NullString

	row := config.DB.QueryRow(`
    SELECT sign, moderated, modsign, ts, type
    FROM msgresult
    WHERE sign = ?;`, cert.Sign)

	err = row.Scan(&sign, &moderated, &modsign, &ts, &existingType)
	if err != nil {
		fmt.Println("Error scanning row:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no record found after insert failed: %w", err)
		}
		return nil, fmt.Errorf("failed to scan existing record: %w", err)
	}

	fmt.Println("Fetching existing record:", sign, moderated, modsign, existingType)

	// If this is an incoming report:
	if cert.Type == "report" {
		// If the existing record was an approved manual_mod message (or if re-reporting an approved message):
		// Reset it back to unmoderated report state so moderators can review it.
		if existingType.String != "report" {
			updateQuery := `
            UPDATE msgresult
            SET type = 'report',
                reason = ?,
                moderated = NULL,
                modsign = NULL
            WHERE sign = ?;`
			if _, err := config.DB.Exec(updateQuery, cert.Reason, cert.Sign); err != nil {
				return nil, fmt.Errorf("failed to update record to report: %w", err)
			}
			fmt.Println("Updated existing message to report successfully")
			return &models.ModResponse{
				Sign:      cert.Sign,
				PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
				Status:    "acknowledged",
			}, nil
		}

		// If it's already a report and this mod has already made a decision:
		if moderated.Valid && modsign.Valid && modsign.String != "" {
			return &models.ModResponse{
				Sign:      modsign.String,
				PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
				Status:    strconv.FormatInt(moderated.Int64, 10),
			}, nil
		}

		// Still unmoderated report awaiting review:
		return &models.ModResponse{
			Sign:      cert.Sign,
			PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
			Status:    "acknowledged",
		}, nil
	}

	// For non-report types (e.g. manual_mod image submissions):
	if moderated.Valid && modsign.Valid && modsign.String != "" {
		return &models.ModResponse{
			Sign:      modsign.String,
			PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
			Status:    strconv.FormatInt(moderated.Int64, 10),
		}, nil
	}

	return &models.ModResponse{
		Sign:      cert.Sign,
		PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
		Status:    "acknowledged",
	}, nil
}

// UpdateModerationStatus updates the moderation status of a message and signs the update
func UpdateModerationStatus(sign string, modsign string, moderated int) (*models.ModResponse, error) {

	// Update the database with moderated status and signature
	updateQuery := `
        UPDATE msgresult
        SET moderated = ?, modsign = ?
        WHERE sign = ?;`

	_, err := config.DB.Exec(updateQuery, moderated, modsign, sign)
	if err != nil {
		return nil, fmt.Errorf("failed to update moderation status: %w", err)
	}
	fmt.Println("Moderation status updated successfully")
	return &models.ModResponse{
		Sign:      modsign,
		PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
		Status:    strconv.Itoa(moderated),
	}, nil
}

// GetUnmoderatedMsgs returns all messages from msgresult where moderated and modsign are NULL
func GetUnmoderatedMsgs() ([]models.MsgCert, error) {
	query := `
        SELECT sign, content, reason, ts, type
        FROM msgresult
        WHERE moderated IS NULL AND modsign IS NULL;`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unmoderated messages: %w", err)
	}

	defer rows.Close()

	var msgs []models.MsgCert
	for rows.Next() {
		var cert models.MsgCert
		var content, reason, msgType string
		var ts int64
		if err := rows.Scan(&cert.Sign, &content, &reason, &ts, &msgType); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		cert.Msg.Content = content
		cert.Reason = reason
		cert.Msg.Ts = ts
		cert.Type = msgType
		fmt.Println("cert", cert)
		msgs = append(msgs, cert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return msgs, nil
}

// GetMsgBySign fetches content, ts, and type for a message from the DB by its sign.
func GetMsgBySign(sign string) (content string, ts int64, msgType string, err error) {
	row := config.DB.QueryRow(`
		SELECT content, ts, COALESCE(type, '') FROM msgresult WHERE sign = ?;`, sign)
	err = row.Scan(&content, &ts, &msgType)
	if err != nil {
		return "", 0, "", fmt.Errorf("GetMsgBySign: %w", err)
	}
	return content, ts, msgType, nil
}

func ReportModSign(cert *types.MsgCert, status string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (string, error) {
	var payload string
	if cert.Type == "manual_mod" {
		// Image messages follow the same structure as auto-mod: content + timestamp + status
		payload = cert.Msg.Content + strconv.FormatInt(cert.Msg.Ts, 10) + status
	} else {
		payload = cert.Sign + status
	}
	_, sign, err := cryptoutils.SignMessage(privateKey, payload)
	if err != nil {
		return "", err
	}
	return sign, nil
}
