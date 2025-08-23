package network

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/node"
	"github.com/libr-forum/Libr/core/db/internal/routing"
)

type PingRequest struct {
	NodeID string `json:"node_id"`
}

type PingResponse struct {
	Status string `json:"status"`
}

type StoredResponse struct {
	Type   string         `json:"type"`
	Status string         `json:"status"`
	Nodes  []*models.Node `json:"nodes"`
}

func HandlePing(body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in HandlePing")
		return nil
	}
	nodeIDStr, ok := bodyMap["node_id"].(string)
	fmt.Println("node_id in handle ping:", nodeIDStr)
	if !ok || nodeIDStr == "" {
		fmt.Println("Missing or invalid node_id in HandlePing")
		return nil
	}
	peerId, ok := bodyMap["peer_id"].(string)
	fmt.Println("peer_id in handle ping:", peerId)
	if !ok || peerId == "" {
		fmt.Println("Missing or invalid peer_id in HandlePing")
		return nil
	}
	nodeID, err := node.DecodeNodeID(nodeIDStr)
	if err != nil {
		fmt.Println("Error decoding node ID:", err)
		return nil
	}

	senderNode := &models.Node{
		NodeId:   nodeID,
		PeerId:   peerId,
		LastSeen: time.Now().Unix(),
	}

	if GlobalPinger == nil {
		fmt.Println("❌ Pinger not registered")
		return nil
	}
	rt.InsertNode(localNode, senderNode, GlobalPinger)
	routing.GlobalRT = rt // Update the global reference

	fmt.Println("Routing Table ppp = ", rt.String())

	fmt.Printf("Ping from node ID: %x, Peer ID: %s\n", nodeID, senderNode.PeerId)
	data, err := json.Marshal(PingResponse{Status: "ok"})
	if err != nil {
		fmt.Println("Error while marshaling the PingResponse: ", err)
	}
	return data
}

func FindValueHandler(key string, localNode *models.Node, rt *routing.RoutingTable) []byte {

	values, closest := SendFindValue(key, localNode, rt)

	if values != nil {
		fmt.Println("Found the value")
		type FoundResponse struct {
			Type   string              `json:"type"`
			Values []models.RetMsgCert `json:"values"`
		}
		resp := FoundResponse{
			Type:   "found",
			Values: values,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error while marshiling the PingResponse: ", err)
		}
		fmt.Println("Data from find value:", string(data))
		return data
	} else {
		fmt.Println("Didn't find the value")
		type RedirectResponse struct {
			Type  string         `json:"type"`
			Nodes []*models.Node `json:"nodes"`
		}
		resp := RedirectResponse{
			Type:  "redirect",
			Nodes: closest,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error while marshiling the PingResponse: ", err)
		}

		return data
	}
}

func FindNodeHandler(body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	fmt.Printf("[DEBUG] find_node_id type: %T, value: %#v\n", bodyMap["find_node_id"], bodyMap["find_node_id"])

	if !ok {
		fmt.Println("Invalid body format in FindNodeHandler")
		return nil
	}

	findNodeStr, ok := bodyMap["find_node_id"].(string)
	if !ok || findNodeStr == "" {
		fmt.Println("Missing or invalid find_node_id in FindNodeHandler")
		return nil
	}

	nodeIDStr, ok := bodyMap["node_id"].(string)
	if !ok || nodeIDStr == "" {
		fmt.Println("Missing or invalid node_id in FindNodeHandler") // fixed message
		return nil
	}

	peerId, ok := bodyMap["peer_id"].(string)
	if !ok || peerId == "" {
		fmt.Println("Missing or invalid peer_id in FindNodeHandler") // fixed message
		return nil
	}

	nodeIDStr = strings.TrimSpace(nodeIDStr) // prevent decode errors from stray spaces
	nodeId, err := node.DecodeNodeID(nodeIDStr)
	if err != nil {
		fmt.Println("Error decoding node ID:", err)
		return nil
	}
	// Generate sender node ID from public key

	senderNode := &models.Node{
		NodeId:   nodeId,
		PeerId:   peerId,
		LastSeen: time.Now().Unix(),
	}

	// Decode the target node ID from hex
	findNodeStr = strings.TrimSpace(findNodeStr)
	decKey, err := node.DecodeNodeID(findNodeStr)

	if err != nil {
		fmt.Println("Error decoding find node ID:", err)
		return nil
	}

	if GlobalPinger == nil {
		fmt.Println("❌ Pinger not registered")
		return nil
	}

	// Lookup closest nodes to the target ID
	closest := SendFindNode(decKey, rt)

	// Insert sender node into routing table
	rt.InsertNode(localNode, senderNode, GlobalPinger)
	routing.GlobalRT = rt // Update the global reference

	fmt.Println("Routing Table lll = ", rt.String())

	data, err := json.Marshal(closest)
	if err != nil {
		fmt.Println("Error while marshaling the PingResponse:", err)
		return nil
	}

	return data
}

func StoreHandler(body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	// Optionally, validate and use public_key for senderNode if needed

	var msgcert models.MsgCert
	jsonBytes, _ := json.Marshal(body)
	if err := json.Unmarshal(jsonBytes, &msgcert); err != nil {
		fmt.Println("Error unmarshaling into MsgCert:", err)
		return nil
	}

	tsmin := msgcert.Msg.Ts
	tsmin = tsmin - (tsmin % 60)
	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest, stored := StoreValue(keyBytes, &msgcert, localNode, rt)

	if !stored {
		fmt.Println("Sending list of k closest nodes")
		type RedirectResponse struct {
			Type  string         `json:"type"`
			Nodes []*models.Node `json:"nodes"`
		}
		resp := RedirectResponse{
			Type:  "redirect",
			Nodes: closest,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error while marshiling the PingResponse: ", err)
		}

		return data
	}

	fmt.Println("Store at: ", localNode)
	resp := StoredResponse{
		Type:   "stored",
		Status: "ok",
		Nodes:  closest,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}
	fmt.Println("data:", string(data))
	return data
}

func DeleteHandler(repCert models.ReportCert, localNode *models.Node, rt *routing.RoutingTable) []byte {
	tsmin := repCert.Msgcert.Msg.Ts
	tsmin -= (tsmin % 60)

	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest, err := DeleteValue(&keyBytes, &repCert, localNode, rt)
	if err != nil {
		fmt.Println("Error", err)
		type ErrorResponse struct {
			Type  string `json:"type"`
			Error string `json:"error"`
		}
		resp := ErrorResponse{
			Type:  "redirect",
			Error: err.Error(),
		}
		data, _ := json.Marshal(resp)
		return data
	}

	if closest != nil {
		fmt.Println("Sending list of k closest nodes")
		type RedirectResponse struct {
			Type  string         `json:"type"`
			Nodes []*models.Node `json:"nodes"`
		}
		resp := RedirectResponse{
			Type:  "redirect",
			Nodes: closest,
		}
		data, _ := json.Marshal(resp)
		return data
	}

	fmt.Println("Store at:", localNode)
	resp := StoredResponse{
		Type:   "deleted",
		Status: "ok",
	}
	data, _ := json.Marshal(resp)
	return data
}
