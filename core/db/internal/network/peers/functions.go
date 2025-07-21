package peer

import (
	// ...
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/network/bootstrap"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/joho/godotenv"
)

var Peer *ChatPeer
var globalLocalNode *node.Node
var globalRT *routing.RoutingTable

func RegisterLocalState(n *node.Node, rt *routing.RoutingTable) {
	globalLocalNode = n
	globalRT = rt

	// ✅ Register POST handler
	network.RegisterPOST(POST)

	// ✅ Register RealPinger
	network.RegisterPinger(&network.RealPinger{})
}

func initDHT() {
	baseEnvPath := os.Getenv("DB_ENV_PATH")
	if baseEnvPath == "" {
		baseEnvPath = "./core/db/.env" // fallback when running outside Docker
	}

	// 1. Load base .env
	baseEnv, err := godotenv.Read(baseEnvPath)
	if err != nil {
		log.Fatal("Failed to read base .env:", err)
	}

	// 2. Parse IP and PORT from publicAddr
	parts := strings.Split(publicAddr, ":")
	if len(parts) != 2 {
		log.Fatalf("Invalid public address format: %s", publicAddr)
	}
	ip := parts[0]
	port := parts[1]
	selfAddr := ip + ":" + port

	// 3. Read and parse current DB_PORT
	dbPortStr := baseEnv["DB_PORT"]
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT in .env: %v", err)
	}

	// 4. Append to BOOTSTRAP in base .env if DB_PORT is a seed
	isSeed := dbPort == 5432 || dbPort == 5440 || dbPort == 5447 || dbPort == 5454 || dbPort == 5463
	if isSeed {
		bootstrapList := strings.Split(baseEnv["BOOTSTRAP"], ",")
		exists := false
		for _, b := range bootstrapList {
			if strings.TrimSpace(b) == selfAddr {
				exists = true
				break
			}
		}
		if !exists {
			if baseEnv["BOOTSTRAP"] == "" {
				baseEnv["BOOTSTRAP"] = selfAddr
			} else {
				baseEnv["BOOTSTRAP"] += "," + selfAddr
			}
			fmt.Println("✅ Added", selfAddr, "to base BOOTSTRAP")
		}
	}

	// 5. Add logic to get bootstrap nodes from csv
	bootstrapAddrs := []string{}
	if baseEnv["BOOTSTRAP"] != "" {
		bootstrapAddrs = strings.Split(baseEnv["BOOTSTRAP"], ",")
	}

	// 6. Init DB and routing
	config.InitConnection(port)
	address := ip + ":" + port
	localNode := &node.Node{
		NodeId: node.GenerateNodeID(address),
		IP:     ip,
		Port:   port,
	}
	rt := routing.GetOrCreateRoutingTable(localNode)
	RegisterLocalState(localNode, rt)

	// 7. Bootstrap to other nodes
	for _, addr := range bootstrapAddrs {
		addr = strings.TrimSpace(addr)
		if addr == "" || addr == selfAddr {
			continue
		}
		targetIP := addr[:strings.LastIndex(addr, ":")]
		targetPort := addr[strings.LastIndex(addr, ":")+1:]
		fmt.Println("Bootstrapping with", addr)
		bootstrap.Bootstrap(targetIP, targetPort, localNode, rt)
	}

	data, _ := json.MarshalIndent(rt, "", "  ")
	fmt.Println(string(data))
	fmt.Println("✅ Kademlia node running at", address)
}

func StartNode(relayAdd string) {

	fmt.Println("Starting Node...")

	relayAddr := relayAdd // have to build logic about getting Multiple Relays

	var err error
	Peer, err = NewChatPeer(relayAddr)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return
	}
	ctx := context.Background()

	if err := Peer.Start(ctx); err != nil {
		fmt.Println(err)
	}

	initDHT()

}

func POST(targetIP, targetPort, route string, body []byte) ([]byte, error) {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reqParams := make(map[string]string)
	parts := strings.Split(route, "/")
	params := strings.Split(parts[1], "&&")

	for _, param := range params {
		keyVal := strings.Split(param, "=")
		if len(keyVal) == 2 {
			reqParams[keyVal[0]] = keyVal[1]
		}
	}
	reqParams["Method"] = "POST"

	jsonReq, err := json.Marshal(reqParams)
	if err != nil {
		fmt.Println("[DEBUG] Failed to marshal POST request params")
		return nil, err
	}

	resp, err := Peer.Send(timeoutCtx, targetIP, targetPort, jsonReq, body)
	if err != nil {
		fmt.Println("[DEBUG] Error sending POST request:", err)
		return nil, err
	}

	return resp, nil
}

func GET(targetIP, targetPort, route string) ([]byte, error) {
	ctx := context.Background()

	reqParams := make(map[string]string)
	parts := strings.Split(route, "/")
	params := strings.Split(parts[1], "&&")

	for _, param := range params {
		keyVal := strings.Split(param, "=")
		if len(keyVal) == 2 {
			reqParams[keyVal[0]] = keyVal[1]
		}
	}
	reqParams["Method"] = "GET"

	jsonReq, err := json.Marshal(reqParams)
	if err != nil {
		fmt.Println("[DEBUG] Failed to marshal GET request params")
		return nil, err
	}

	resp, err := Peer.Send(ctx, targetIP, targetPort, jsonReq, nil)
	if err != nil {
		fmt.Println("[DEBUG] Error sending GET request:", err)
		return nil, err
	}

	return resp, nil
}

func ServeGetReq(paramsBytes []byte) []byte {
	var params map[string]interface{}
	err := json.Unmarshal(paramsBytes, &params)
	if err != nil {
		fmt.Println(err)
	}

	switch params["route"] {
	case "find_value":
		keyStr, ok := params["ts"].(string)
		if !ok {
			fmt.Println("ts is not a string")
		}
		fmt.Printf("Timestamp to retrieve: %s", keyStr)
		return network.FindValueHandler(keyStr, globalLocalNode, globalRT)
	}

	var resp []byte
	return resp

}

func ServePostReq(paramsBytes []byte, bodyBytes []byte) []byte {
	fmt.Println("Serving Post Request")
	var params map[string]interface{}
	err := json.Unmarshal(paramsBytes, &params)
	if err != nil {
		fmt.Println(err)
	}
	ip := strings.Split("49.36.179.166:50613", ":")[0]

	var body map[string]interface{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Body", body)

	switch params["route"] {
	case "ping":
		return network.HandlePing(ip, body, globalLocalNode, globalRT)

	case "store":
		var msgCert models.MsgCert
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Println("Error re-marshaling body map:", err)
		}
		err = json.Unmarshal(jsonBytes, &msgCert)
		if err != nil {
			fmt.Println("Error unmarshaling into MsgCert:", err)
		}
		return network.StoreHandler(msgCert, globalLocalNode, globalRT)
	case "find_node":
		var body map[string]interface{}
		err := json.Unmarshal(bodyBytes, &body)
		if err != nil {
			fmt.Println(err)
		}

		keyStr, ok := body["node_id"].(string)
		if !ok {
			fmt.Println("node_id is not a string")
		}

		// Assuming the key string is a hex-encoded 20-byte ID:
		decodedKey, err := node.DecodeNodeID(keyStr)
		if err != nil {
			fmt.Println("failed to decode node ID:", err)
		}
		return network.FindNodeHandler(decodedKey, globalLocalNode, globalRT)
	}
	var PostResp []byte
	return PostResp
}
