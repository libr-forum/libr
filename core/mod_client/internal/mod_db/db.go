package moddb

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

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
    INSERT INTO msgresult (sign, content, reason)
    VALUES (?, ?, ?);`

	_, err := config.DB.Exec(insertQuery, cert.Sign, cert.Msg.Content, cert.Reason)
	if err == nil {
		fmt.Println("Message result stored successfully")
		return &models.ModResponse{
			Sign:      cert.Sign,
			PublicKey: "",
			Status:    "acknowledged",
		}, nil
	}
	log.Printf("Insert failed")
	var moderated sql.NullInt64
	var modsign sql.NullString
	var sign string

	if rand.Intn(2) == 0 {
		fmt.Println("Running the Test logic...")
		TestManModerateMsg(cert)
		// Put your actual code here
	} else {
		fmt.Println("Skipping this time.")
	}

	row := config.DB.QueryRow(`
    SELECT sign, moderated, modsign
    FROM msgresult
    WHERE sign = ?;`, cert.Sign)

	err = row.Scan(&sign, &moderated, &modsign)
	if err != nil {
		fmt.Println("Error scanning row:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no record found after insert failed: %w", err)
		}
		return nil, fmt.Errorf("failed to scan existing record: %w", err)
	}

	fmt.Println("Fetching existing record:", sign, moderated, modsign)

	// Only proceed if moderated is non-NULL and equals 1
	if moderated.Valid && moderated.Int64 == 1 && modsign.Valid && modsign.String != "" {
		// payload := fmt.Sprintf("%d", moderated.Int64) + modsign.String

		// pub, priv, err := cryptoutils.LoadKeys()
		// if err != nil {
		// 	log.Printf("Key load error: %v", err)
		// 	return nil, fmt.Errorf("failed to load keys: %w", err)
		// }

		// _, signed, err := cryptoutils.SignMessage(priv, payload)
		// if err != nil {
		// 	return nil, fmt.Errorf("signing sign-moderated: %w", err)
		// }

		return &models.ModResponse{
			Sign:      modsign.String,
			PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
			Status:    strconv.FormatInt(moderated.Int64, 10),
		}, nil
	}

	return &models.ModResponse{
		Sign:      cert.Sign,
		PublicKey: cert.PublicKey,
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
		fmt.Println("cert", cert)
		msgs = append(msgs, cert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return msgs, nil
}

func TestManModerateMsg(cert types.MsgCert) {
	fmt.Println("Testing manual moderation for message:", cert.Sign)
	rand.Seed(time.Now().UnixNano()) // seed with current time
	status := rand.Intn(2)
	modsign, _ := ReportModSign(&cert, strconv.Itoa(status), keycache.PrivKey, keycache.PubKey)
	fmt.Println("Mod signature:", modsign)
	UpdateModerationStatus(cert.Sign, modsign, status)
}

func ReportModSign(cert *types.MsgCert, status string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (string, error) {

	payload := cert.Sign + status
	_, sign, err := cryptoutils.SignMessage(privateKey, payload)
	if err != nil {
		return "", err
	}

	return sign, nil
}
