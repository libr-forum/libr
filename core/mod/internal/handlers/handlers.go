package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

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

var (
	msgStore = make(map[string]models.ModResponse)
	mu       sync.RWMutex
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to LIBR prototype"))
}

func MsgIN(w http.ResponseWriter, r *http.Request) {
	var req models.UserMsg
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid user message: %v", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, "content missing", http.StatusBadRequest)
		return
	}

	// Moderate message
	moderationStatus, err := service.ModerateMsg(req)
	if err != nil {
		log.Printf("Moderation error: %v", err)
		http.Error(w, "error during moderation", http.StatusInternalServerError)
		return
	}

	// Load keys to sign
	pub, priv, err := cryptoutils.LoadKeys()
	if err != nil {
		log.Printf("Key load error: %v", err)
		http.Error(w, "failed to load keys", http.StatusInternalServerError)
		return
	}

	// Sign
	signed, err := service.ModSign(req, moderationStatus, priv, pub)
	if err != nil {
		log.Printf("Signing error: %v", err)
		http.Error(w, "error signing message", http.StatusInternalServerError)
		return
	}
	var response models.ModResponse
	json.Unmarshal([]byte(signed), &response)
	fmt.Println(response)
	// Respond
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
