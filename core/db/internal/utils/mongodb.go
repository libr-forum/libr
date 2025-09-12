package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
)

var MongoClient *mongo.Client


type modResp struct {
	ModLists []models.Mods `json:"modlist"`
}


type NodeResp struct {
	NodesLists []models.Node `json:"nodeslist"`
}


func GetRelayAddrFromJSServer() ([]string, error) {
	cf, err := config.ReadDBConfigFile()
	if err != nil {
		fmt.Println("Error reading dbconfig.json:", err)
		return nil, err
	}
	serverURL := cf.JS_ServerURL

	req, err := http.NewRequest("GET", serverURL+"/api/getrelay", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	var payload relayResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}
	fmt.Println(resp.Body)
	fmt.Println(payload)

	var addresses []string
for _, relay := range payload.RelayList.Relays {
    addresses = append(addresses, relay.Address)
}
	fmt.Println(addresses)
	return addresses, nil
}


type relayResp struct {
    RelayList struct {
        Relays []struct {
            Address string `json:"address"`
        } `json:"relaylist"`
    } `json:"relay_list"`
}

// ✅ Fetch mods
func GetModsFromJSServer() ([]*models.Mods, error) {
	cf, err := config.ReadDBConfigFile()
	if err != nil {
		fmt.Println("Error reading dbconfig.json:", err)
		return nil, err
	}
	serverURL := cf.JS_ServerURL

	req, err := http.NewRequest("GET", serverURL+"/api/getmod", nil) // 🔥 corrected endpoint
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var ModReturnList []*models.Mods

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	var payload modResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	for i := range payload.ModLists {
		m := payload.ModLists[i] // copy to avoid pointer bug
		ModReturnList = append(ModReturnList, &m)
	}

	return ModReturnList, nil
}

// ✅ Fetch DB nodes
func GetDBFromJSServer() ([]*models.Node, error) {
	cf, err := config.ReadDBConfigFile()
	if err != nil {
		fmt.Println("Error reading dbconfig.json:", err)
		return nil, err
	}
	serverURL := cf.JS_ServerURL

	req, err := http.NewRequest("GET", serverURL+"/api/getboot", nil) // 🔥 corrected endpoint
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var NodeReturnList []*models.Node

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	var payload NodeResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	for i := range payload.NodesLists {
		n := payload.NodesLists[i] // copy to avoid pointer bug
		NodeReturnList = append(NodeReturnList, &n)
	}

	return NodeReturnList, nil
}
