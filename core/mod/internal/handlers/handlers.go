package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod/internal/service"
	"github.com/devlup-labs/Libr/core/mod/models"
)

func HandleMsg() {
	// 1. msg in
	// 2. validate
	// 3. moderate
	// 4. sign
	// 5. respond
}

func MsgIN(bodyBytes []byte) []byte {
	var req models.UserMsg

	err := json.Unmarshal(bodyBytes, &req)
	if err != nil {
		fmt.Println("Invalid JSON")
		return nil
	}

	// Moderate message
	moderationStatus, err := service.ModerateMsg(req)
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

	return []byte(signed)
}

// func MsgOUT(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	timestamp := vars["timestamp"]

// 	mu.RLock()
// 	msg, exists := msgStore[timestamp]
// 	mu.RUnlock()

// 	if !exists {
// 		http.Error(w, "message not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(msg)
// }
