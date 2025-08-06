package moddb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod_client/config"
	"github.com/devlup-labs/Libr/core/mod_client/models"
	"github.com/devlup-labs/Libr/core/mod_client/types"
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
	insertQuery := `
    INSERT INTO msgresult (sign, content, reason)
    VALUES (?, ?, ?);`

	_, err := config.DB.Exec(insertQuery, cert.Sign, cert.Msg.Content, cert.Reason)
	if err == nil {
		return &models.ModResponse{
			Sign:      cert.Sign,
			PublicKey: "",
			Status:    "acknowledged",
		}, nil
	}

	var moderated int
	var modsign string
	var sign string

	row := config.DB.QueryRow(`
        SELECT sign, moderated, modsign
        FROM msgresult
        WHERE sign = ?;`, cert.Sign)

	err = row.Scan(&sign, &moderated, &modsign)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no record found after insert failed: %w", err)
		}
		return nil, fmt.Errorf("failed to scan existing record: %w", err)
	}

	if moderated == 1 && modsign != "" {
		payload := fmt.Sprintf("%d", moderated) + modsign

		pub, priv, err := cryptoutils.LoadKeys()
		if err != nil {
			log.Printf("Key load error: %v", err)
			return nil, fmt.Errorf("failed to load keys: %w", err)
		}

		_, signed, err := cryptoutils.SignMessage(priv, payload)
		if err != nil {
			return nil, fmt.Errorf("signing sign-moderated: %w", err)
		}

		return &models.ModResponse{
			Sign:      sign,
			PublicKey: string(pub),
			Status:    signed,
		}, nil
	}

	return &models.ModResponse{
		Sign:      cert.Sign,
		PublicKey: cert.PublicKey,
		Status:    "acknowledged",
	}, nil
}

// UpdateModerationStatus updates the moderation status of a message and signs the update
func UpdateModerationStatus(sign string, moderated int) (*models.ModResponse, error) {
	// Prepare payload for signing
	payload := fmt.Sprintf("%d", moderated) + sign

	pub, priv, err := cryptoutils.LoadKeys()
	if err != nil {
		log.Printf("Key load error: %v", err)
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}

	_, signature, err := cryptoutils.SignMessage(priv, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign moderation payload: %w", err)
	}

	// Update the database with moderated status and signature
	updateQuery := `
        UPDATE msgresult
        SET moderated = ?, modsign = ?
        WHERE sign = ?;`

	_, err = config.DB.Exec(updateQuery, moderated, signature, sign)
	if err != nil {
		return nil, fmt.Errorf("failed to update moderation status: %w", err)
	}

	return &models.ModResponse{
		Sign:      sign,
		PublicKey: string(pub),
		Status:    signature,
	}, nil
}

// GetUnmoderatedMsgs returns all messages from msgresult where moderated and modsign are NULL
func GetUnmoderatedMsgs() ([]models.MsgCert, error) {
	query := `
        SELECT sign, content, reason
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
		var content, reason string
		if err := rows.Scan(&cert.Sign, &content, &reason); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		cert.Msg.Content = content
		cert.Reason = reason
		msgs = append(msgs, cert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return msgs, nil
}
