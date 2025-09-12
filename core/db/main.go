package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/keycache"
	peer "github.com/libr-forum/Libr/core/db/internal/network/peers"
	"github.com/libr-forum/Libr/core/db/internal/node"
	"github.com/libr-forum/Libr/core/db/internal/routing"
	"github.com/libr-forum/Libr/core/db/internal/utils"
)

var JS_API_key string
var JS_ServerURL string

func main() {
	keycache.InitKeys()
	godotenv.Load()
	   err := CreateDbConfig()
	   if err != nil {
		   fmt.Println("Error creating config file. Kindly create your own if required")
	   }
	   cf, err := config.ReadDBConfigFile()
	   if err != nil {
		   fmt.Println("Error reading values from config file")
	   }
	   JS_API_key = cf.API_KEY

	   JS_ServerURL = cf.JS_ServerURL
	   
	//utils.SetupMongo("mongodb+srv://peer:peerhehe@cluster0.vswojqe.mongodb.net/")
	//utils.SetupMongo(JS_ServerURL)
	relayAddrs, err := utils.GetRelayAddrFromJSServer()
	
	if err != nil {
		fmt.Println("Error while getting relay address, ", err)
	}

	fmt.Println(relayAddrs)

	go func() {
    for {
        // Close old host if it exists
        if peer.Peer != nil && peer.Peer.Host != nil {
            fmt.Println("[DEBUG] Closing host")
            peer.Peer.Host.Close()
        }

        // Start a new node right away
        fmt.Println("[DEBUG] Starting new node...")
        peer.StartNode(relayAddrs)

        // ⏰ Calculate how long to wait until the next o'clock
        time.Sleep(55*time.Minute)
    }
}()


	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Interrupt received. Exiting gracefully.")
	if(config.DBtype=="boot"){
	deleteFromJSServer()
	}
	fmt.Println("Sent delete request to JS server")
	if routing.GlobalRT != nil {
		routing.GlobalRT.SaveToDBAsync()
		time.Sleep(1 * time.Second)
	}
}


func deleteFromJSServer() error {
	deleteData := map[string]string{
		"node_id" : node.GenerateNodeIDFromPublicKey(),
	}

	jsonData, err := json.Marshal(deleteData)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("DELETE", JS_ServerURL+"/api/deleteboot", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", JS_API_key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	fmt.Printf("Successfully deleted relay")
	return nil
}

func CreateDbConfig() error {
	   path := config.GetDBConfigPath()

	   if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		   return fmt.Errorf("failed to create modconfig directory: %w", err)
	   }

	   // Only create the file if it does not already exist
	   if _, err := os.Stat(path); os.IsNotExist(err) {
		   f, err := os.Create(path)
		   if err != nil {
			   return fmt.Errorf("failed to create config file: %w", err)
		   }
		   defer f.Close()

		   // Write an empty JSON object if file is new
		   _, err = f.WriteString("{}")
		   if err != nil {
			   return fmt.Errorf("failed to write empty JSON to config file: %w", err)
		   }
	   }
	   return nil
}
