package peer

import (
	// ...
	"bytes"
	"context"
	"errors"

	//"encoding/base64"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/keycache"
	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/network"
	"github.com/libr-forum/Libr/core/db/internal/network/bootstrap"
	"github.com/libr-forum/Libr/core/db/internal/node"
	"github.com/libr-forum/Libr/core/db/internal/routing"
	"github.com/libr-forum/Libr/core/db/internal/utils"
)

var Peer *ChatPeer
var globalLocalNode *models.Node

type RelayDist struct {
	relayID string
	dist    *big.Int
}

func RegisterLocalState(n *models.Node, rt *routing.RoutingTable) {
	globalLocalNode = n
	routing.GlobalRT = rt

	// ✅ Register POST handler
	network.RegisterPOST(POST)

	// ✅ Register RealPinger
	network.RegisterPinger(&network.RealPinger{})
}

func initDHT() {
	bootstrapAddrs, _ := utils.GetDBFromJSServer()

	// fmt.Print("Bootstrap addresses: ", bootstrapAddrs, "\n")

	// 3. Init DB and routing
	config.InitDB()
	nodeId := node.GenerateNodeID(base64.StdEncoding.EncodeToString(keycache.PubKey))
	localNode := &models.Node{
		NodeId: nodeId,
		PeerId: PeerID,
	}
	nodeIdStr := base64.StdEncoding.EncodeToString(nodeId[:])
	fmt.Println("Node ID:", nodeIdStr)
	rt := routing.GetOrCreateRoutingTable(localNode.NodeId)
	RegisterLocalState(localNode, rt)

	fmt.Println("🌐 Starting Kademlia node at")

	empty := true
	for _, b := range rt.Buckets {
		if b != nil && len(b.Nodes) > 0 {
			empty = false
			break
		}
	}
	if empty {
		fmt.Println("❗ No buckets found in routing table, bootstrapping from peers...")
		bootstrap.BootstrapFromPeers(bootstrapAddrs, localNode, rt)
	} else {
		// pass bootstrapAddrs for fallback
		bootstrap.NodeUpdate(localNode, rt, bootstrapAddrs)
	}

	data, _ := json.MarshalIndent(rt, "", "  ")
	fmt.Println(string(data))
	fmt.Println("Peer ID:", PeerID)
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

func GET(targetPeerID string, route string) ([]byte, error) { //"/ts=123&&id=123"

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

	GetResp, err := Peer.Send(ctx, targetPeerID, jsonReq, nil)
	if err != nil {
		fmt.Println("Error Sending trial get message")
	}
	GetResp = bytes.TrimRight(GetResp, "\x00")
	return GetResp, nil //this will be json bytes with resp encoded in form of resp from the server and can be used according to utility
}

func POST(targetPeerID string, route string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	parts := strings.SplitN(route, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid route format: %s", route)
	}

	reqparams := make(map[string]string)
	for _, param := range strings.Split(parts[1], "&&") {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) == 2 {
			reqparams[kv[0]] = kv[1]
		}
	}
	reqparams["Method"] = "POST"

	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal req params: %w", err)
	}

	GetResp, err := Peer.Send(ctx, targetPeerID, jsonReq, body)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("⏳ POST request timed out after 5s")
		} else {
			fmt.Println("❌ POST request failed:", err)
		}
		return nil, err
	}
	if bytes.Equal(GetResp, []byte("Target peer not found")) || GetResp == nil || len(GetResp) == 0 {
		return nil, errors.New("ping failed: empty response")
	}

	return bytes.TrimRight(GetResp, "\x00"), nil
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
		return network.FindValueHandler(keyStr, globalLocalNode, routing.GlobalRT)
	}

	var resp []byte
	return resp

}

func ServePostReq(peerId string, paramsBytes []byte, bodyBytes []byte) []byte {
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

	fmt.Println("Peer ID:", peerId)

	switch route {
	case "ping":
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			fmt.Println("Failed to unmarshal body:", err)
			return nil
		}
		return network.HandlePing(body, globalLocalNode, routing.GlobalRT)

	case "store":
		var msgCert models.MsgCert
		if err := json.Unmarshal(bodyBytes, &msgCert); err != nil {
			fmt.Println("Error unmarshaling into MsgCert:", err)
			return nil
		}
		return network.StoreHandler(msgCert, globalLocalNode, routing.GlobalRT)

	case "find_node":
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			fmt.Println("Failed to unmarshal body:", err)
			return nil
		}
		return network.FindNodeHandler(body, globalLocalNode, routing.GlobalRT)

	case "delete":
		var repCert models.ReportCert
		if err := json.Unmarshal(bodyBytes, &repCert); err != nil {
			fmt.Println("Error unmarshaling into ReportCert:", err)
			return nil
		}
		return network.DeleteHandler(repCert, globalLocalNode, routing.GlobalRT)

	default:
		fmt.Println("Unknown POST route:", route)
		return nil
	}
}

// package peer

// import (
// 	// ...
// 	"bytes"
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"math/big"
// 	"strings"
// 	"time"

// 	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
// 	"github.com/libr-forum/Libr/core/db/config"
// 	"github.com/libr-forum/Libr/core/db/internal/models"
// 	"github.com/libr-forum/Libr/core/db/internal/network"
// 	"github.com/libr-forum/Libr/core/db/internal/network/bootstrap"
// 	"github.com/libr-forum/Libr/core/db/internal/node"
// 	"github.com/libr-forum/Libr/core/db/internal/routing"
// 	"github.com/libr-forum/Libr/core/db/internal/utils"
// )

// var Peer *ChatPeer
// var globalLocalNode *models.Node
// var GlobalRT *routing.RoutingTable

// type RelayDist struct {
// 	relayID string
// 	dist    *big.Int
// }

// func RegisterLocalState(n *models.Node, rt *routing.RoutingTable) {
// 	globalLocalNode = n
// 	GlobalRT = rt

// 	// ✅ Register POST handler
// 	network.RegisterPOST(POST)

// 	// ✅ Register RealPinger
// 	network.RegisterPinger(&network.RealPinger{})
// }

// func initDHT() {
// 	parts := strings.Split(OwnPubIP, ":")
// 	pubKey, _, _ := cryptoutils.LoadKeys()
// 	if len(parts) != 2 {
// 		log.Fatalf("Invalid public address format: %s", OwnPubIP)
// 	}
// 	ip := parts[0]
// 	port := parts[1]

// 	bootstrapAddrs, _ := utils.GetDbAddr()

// 	// 3. Init DB and routing
// 	config.InitDB()
// 	address := ip + ":" + port
// 	localNode := &models.Node{
// 		NodeId:    node.GenerateNodeID(base64.StdEncoding.EncodeToString(pubKey)),
// 		IP:        ip,
// 		Port:      port,
// 		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
// 	}
// 	rt := routing.GetOrCreateRoutingTable(localNode)
// 	RegisterLocalState(localNode, rt)

// 	fmt.Println("🌐 Starting Kademlia node at")

// 	empty := true
// 	for _, b := range rt.Buckets {
// 		if b != nil && len(b.Nodes) > 0 {
// 			empty = false
// 			break
// 		}
// 	}
// 	if empty {
// 		// All buckets are nil
// 		fmt.Println("❗ No buckets found in routing table, bootstrapping from peers...")
// 		bootstrap.BootstrapFromPeers(bootstrapAddrs, localNode, rt)
// 	} else {
// 		bootstrap.NodeUpdate(rt)
// 	}

// 	data, _ := json.MarshalIndent(rt, "", "  ")
// 	fmt.Println(string(data))
// 	fmt.Println("✅ Kademlia node running at", address)
// }

// func StartNode(relayMultiAddrList []string) {

// 	fmt.Println("Starting Node...")
// 	var err error
// 	Peer, err = NewChatPeer(relayMultiAddrList)
// 	if err != nil {
// 		fmt.Println("Error creating peer:", err)
// 		return
// 	}

// 	ctx := context.Background()

// 	if err := Peer.Start(ctx); err != nil {
// 		log.Fatal(err)
// 	}

// 	initDHT()
// }

// func GET(targetIP string, targetPort string, route string) ([]byte, error) { //"/ts=123&&id=123"

// 	reqparams := make(map[string]string)
// 	parts := strings.Split(route, "/")

// 	params := strings.Split(parts[1], "&&")

// 	for i := range len(params) {
// 		key := strings.Split(params[i], "=")[0]
// 		value := strings.Split(params[i], "=")[1]

// 		reqparams[key] = value
// 	}
// 	reqparams["Method"] = "GET"
// 	jsonReq, err := json.Marshal(reqparams)
// 	if err != nil {
// 		fmt.Println("[DEBUG]Failed to get req params json")
// 		return nil, err
// 	}
// 	_ = jsonReq
// 	ctx := context.Background()

// 	GetResp, err := Peer.Send(ctx, targetIP, targetPort, jsonReq, nil)
// 	if err != nil {
// 		fmt.Println("Error Sending trial get message")
// 	}
// 	GetResp = bytes.TrimRight(GetResp, "\x00")
// 	return GetResp, nil //this will be json bytes with resp encoded in form of resp from the server and can be used according to utility
// }

// func POST(targetIP string, targetPort string, route string, body []byte) ([]byte, error) {

// 	ctx := context.Background()
// 	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	reqparams := make(map[string]string)
// 	parts := strings.Split(route, "/")
// 	params := strings.Split(parts[1], "&&")
// 	for i := range len(params) {
// 		key := strings.Split(params[i], "=")[0]
// 		value := strings.Split(params[i], "=")[1]

// 		reqparams[key] = value
// 	}
// 	reqparams["Method"] = "POST"

// 	jsonReq, err := json.Marshal(reqparams)
// 	if err != nil {
// 		fmt.Println("[DEBUG]Failed to get req params json")
// 		return nil, err
// 	}

// 	GetResp, err := Peer.Send(timeoutCtx, targetIP, targetPort, jsonReq, body)

// 	if err != nil {
// 		fmt.Println("Error Sending trial get message")
// 	}
// 	GetResp = bytes.TrimRight(GetResp, "\x00")
// 	return GetResp, nil
// }

// func ServeGetReq(paramsBytes []byte) []byte {
// 	var params map[string]interface{}
// 	err := json.Unmarshal(paramsBytes, &params)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	switch params["route"] {
// 	case "find_value":
// 		keyStr, ok := params["ts"].(string)
// 		if !ok {
// 			fmt.Println("ts is not a string")
// 		}
// 		fmt.Printf("Timestamp to retrieve: %s", keyStr)
// 		return network.FindValueHandler(keyStr, globalLocalNode, GlobalRT)
// 	}

// 	var resp []byte
// 	return resp

// }

// func ServePostReq(addr []byte, paramsBytes []byte, bodyBytes []byte) []byte {
// 	fmt.Println("Serving Post Request")

// 	var params map[string]interface{}
// 	if err := json.Unmarshal(paramsBytes, &params); err != nil {
// 		fmt.Println("Failed to unmarshal params:", err)
// 		return nil
// 	}

// 	route, ok := params["route"].(string)
// 	if !ok {
// 		fmt.Println("route param missing or not string")
// 		return nil
// 	}

// 	pubipStr := string(addr)
// 	ip := strings.Split(pubipStr, ":")[0]
// 	port := strings.Split(pubipStr, ":")[1]
// 	fmt.Println("IP:", ip, "Port:", port)

// 	var body map[string]interface{}
// 	if err := json.Unmarshal(bodyBytes, &body); err != nil {
// 		fmt.Println("Failed to unmarshal body:", err)
// 		return nil
// 	}
// 	fmt.Println("Body:", body)

// 	switch route {
// 	case "ping":
// 		return network.HandlePing(ip, port, body, globalLocalNode, GlobalRT)

// 	case "store":
// 		var msgCert models.MsgCert
// 		jsonBytes, _ := json.Marshal(body)
// 		if err := json.Unmarshal(jsonBytes, &msgCert); err != nil {
// 			fmt.Println("Error unmarshaling into MsgCert:", err)
// 			return nil
// 		}
// 		return network.StoreHandler(msgCert, globalLocalNode, GlobalRT)

// 	case "find_node":
// 		keyStr, ok := body["node_id"].(string)
// 		if !ok || keyStr == "" {
// 			fmt.Println("find_node error: node_id is missing or not a string")
// 			errResp := map[string]interface{}{"error": "node_id is missing or not a string"}
// 			resp, _ := json.Marshal(errResp)
// 			return resp
// 		}
// 		keyPubKeyStr, ok := body["public_key"].(string)
// 		if !ok || keyPubKeyStr == "" {
// 			fmt.Println("find_node error: public_key is missing or not a string")
// 			errResp := map[string]interface{}{"error": "public_key is missing or not a string"}
// 			resp, _ := json.Marshal(errResp)
// 			return resp
// 		}
// 		// Compose a body map as expected by FindNodeHandler
// 		findNodeBody := map[string]interface{}{
// 			"node_id":    keyStr,
// 			"public_key": keyPubKeyStr,
// 		}
// 		return network.FindNodeHandler(ip, port, findNodeBody, globalLocalNode, GlobalRT)

// 	case "delete":
// 		var repCert models.ReportCert
// 		jsonBytes, _ := json.Marshal(body)
// 		if err := json.Unmarshal(jsonBytes, &repCert); err != nil {
// 			fmt.Println("Error unmarshaling into ReportCert:", err)
// 			return nil
// 		}
// 		return network.DeleteHandler(&repCert, globalLocalNode, GlobalRT)

// 	default:
// 		fmt.Println("Unknown POST route:", route)
// 		return nil
// 	}
// }
