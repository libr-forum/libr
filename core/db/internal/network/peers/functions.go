package peer

import (
	// ...
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/network/bootstrap"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/utils"
)

var Peer *ChatPeer
var globalLocalNode *node.Node
var globalRT *routing.RoutingTable

type RelayDist struct {
	relayID string
	dist    *big.Int
}

func RegisterLocalState(n *node.Node, rt *routing.RoutingTable) {
	globalLocalNode = n
	globalRT = rt

	// ✅ Register POST handler
	network.RegisterPOST(POST)

	// ✅ Register RealPinger
	network.RegisterPinger(&network.RealPinger{})
}

func initDHT() {
	// 1. Parse IP and PORT from publicAddr
	// 1. Parse IP and PORT from publicAddr
	parts := strings.Split(OwnPubIP, ":")
	if len(parts) != 2 {
		log.Fatalf("Invalid public address format: %s", OwnPubIP)
	}
	ip := parts[0]
	port := parts[1]

	// 2.Bootstrap nodes from csv
	start := time.Now()
	bootstrapAddrs, _ := utils.GetDbData()
	elapsed := time.Since(start)
	fmt.Printf("⏱️ utils.GetDBAddrList() took %s\n", elapsed)

	// 3. Init DB and routing
	config.InitDB()
	address := ip + ":" + port
	localNode := &node.Node{
		NodeId: node.GenerateNodeID(address),
		IP:     ip,
		Port:   port,
	}
	rt := routing.GetOrCreateRoutingTable(localNode)
	RegisterLocalState(localNode, rt)

	// 4. Bootstrap to other nodes
	bootstrap.BootstrapFromPeers(bootstrapAddrs, localNode, rt)
	// for _, node := range bootstrapAddrs {
	// 	fmt.Println("Bootstrapping with", node.IP, node.Port)
	// 	bootstrap.Bootstrap(node.IP, node.Port, localNode, rt)
	// }
	// 4. Bootstrap to other nodes
	bootstrap.BootstrapFromPeers(bootstrapAddrs, localNode, rt)
	// for _, node := range bootstrapAddrs {
	// 	fmt.Println("Bootstrapping with", node.IP, node.Port)
	// 	bootstrap.Bootstrap(node.IP, node.Port, localNode, rt)
	// }

	data, _ := json.MarshalIndent(rt, "", "  ")
	fmt.Println(string(data))
	fmt.Println("✅ Kademlia node running at", address)
}

func StartNode(relayMultiAddrList []string) {

	fmt.Println("Starting Node...")
	var err error
	Peer, err = NewChatPeer(relayMultiAddrList)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return
	}

	ctx := context.Background()

	if err := Peer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	initDHT()
}

func GET(targetIP string, targetPort string, route string) ([]byte, error) { //"/ts=123&&id=123"

	reqparams := make(map[string]string)
	parts := strings.Split(route, "/")

	params := strings.Split(parts[1], "&&")

	for i := range len(params) {
		key := strings.Split(params[i], "=")[0]
		value := strings.Split(params[i], "=")[1]

		reqparams[key] = value
	}
	reqparams["Method"] = "GET"
	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		fmt.Println("[DEBUG]Failed to get req params json")
		return nil, err
	}
	_ = jsonReq
	ctx := context.Background()

	GetResp, err := Peer.Send(ctx, targetIP, targetPort, jsonReq, nil)
	if err != nil {
		fmt.Println("Error Sending trial get message")
	}
	GetResp = bytes.TrimRight(GetResp, "\x00")
	return GetResp, nil //this will be json bytes with resp encoded in form of resp from the server and can be used according to utility
}

func POST(targetIP string, targetPort string, route string, body []byte) ([]byte, error) {

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reqparams := make(map[string]string)
	parts := strings.Split(route, "/")
	params := strings.Split(parts[1], "&&")
	for i := range len(params) {
		key := strings.Split(params[i], "=")[0]
		value := strings.Split(params[i], "=")[1]

		reqparams[key] = value
	}
	reqparams["Method"] = "POST"

	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		fmt.Println("[DEBUG]Failed to get req params json")
		return nil, err
	}

	GetResp, err := Peer.Send(timeoutCtx, targetIP, targetPort, jsonReq, body)

	if err != nil {
		fmt.Println("Error Sending trial get message")
	}
	GetResp = bytes.TrimRight(GetResp, "\x00")
	return GetResp, nil
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

func ServePostReq(addr []byte, paramsBytes []byte, bodyBytes []byte) []byte {
	fmt.Println("Serving Post Request")

	var params map[string]interface{}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		fmt.Println("Failed to unmarshal params:", err)
		return nil
	}

	route, ok := params["route"].(string)
	if !ok {
		fmt.Println("route param missing or not string")
		return nil
	}

	pubipStr := string(addr)
	ip := strings.Split(pubipStr, ":")[0]

	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		fmt.Println("Failed to unmarshal body:", err)
		return nil
	}
	fmt.Println("Body:", body)

	switch route {
	case "ping":
		return network.HandlePing(ip, body, globalLocalNode, globalRT)

	case "store":
		var msgCert models.MsgCert
		jsonBytes, _ := json.Marshal(body)
		if err := json.Unmarshal(jsonBytes, &msgCert); err != nil {
			fmt.Println("Error unmarshaling into MsgCert:", err)
			return nil
		}
		return network.StoreHandler(msgCert, globalLocalNode, globalRT)

	case "find_node":
		keyStr, ok := body["node_id"].(string)
		if !ok {
			fmt.Println("node_id is not a string")
			return nil
		}
		decodedKey, err := node.DecodeNodeID(keyStr)
		if err != nil {
			fmt.Println("failed to decode node ID:", err)
			return nil
		}
		return network.FindNodeHandler(decodedKey, globalLocalNode, globalRT)

	case "delete":
		var repCert models.ReportCert
		jsonBytes, _ := json.Marshal(body)
		if err := json.Unmarshal(jsonBytes, &repCert); err != nil {
			fmt.Println("Error unmarshaling into ReportCert:", err)
			return nil
		}
		return network.DeleteHandler(&repCert, globalLocalNode, globalRT)

	default:
		fmt.Println("Unknown POST route:", route)
		return nil
	}
}
