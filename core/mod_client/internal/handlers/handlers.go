package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod_client/internal/service"
	"github.com/devlup-labs/Libr/core/mod_client/models"
	"github.com/devlup-labs/Libr/core/mod_client/types"
)

func MsgIN(bodyBytes []byte) []byte {
	var req models.UserMsg
	err := json.Unmarshal(bodyBytes, &req)
	if err != nil {
		fmt.Println("Invalid JSON")
		return nil
	}

	// Moderate message
	moderationStatus, err := service.AutoModerateMsg(req)
	fmt.Println(moderationStatus)
	if err != nil {
		log.Printf("Moderation error: %v", err)
		return nil
	}

	// Load keys to sign
	pub, priv, err := cryptoutils.LoadKeys()
	if err != nil {
		log.Printf("Key load error: %v", err)
		return nil
	}

	// Sign
	signed, err := service.ModSign(req, moderationStatus, priv, pub)
	if err != nil {

		log.Printf("Signing error: %v", err)
		return nil
	}

	// ✅ Save log for this mod
	_ = service.AppendToModLog(req, moderationStatus)
	return []byte(signed)
}

func MsgReport(bodyBytes []byte) []byte {
	var req types.MsgCert
	err := json.Unmarshal(bodyBytes, &req)
	if err != nil {
		fmt.Println("Invalid JSON")
		return nil
	}

	// Moderate message
	moderationStatus, err := service.ManModerateMsg(req)
	fmt.Println(moderationStatus)
	if err != nil {
		log.Printf("Moderation error: %v", err)
		return nil
	}

	// // ✅ Save log for this mod
	// _ = service.AppendToModLog(req, moderationStatus) for now

	// Return the ModResponse as JSON
	respBytes, err := json.Marshal(moderationStatus)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return nil
	}
	fmt.Println("Returning moderation response:", string(respBytes))
	return respBytes
}
