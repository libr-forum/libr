package network

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

type PingRequest struct {
	NodeID string `json:"node_id"`
}

type PingResponse struct {
	Status string `json:"status"`
}

type StoredResponse struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

func HandlePing(ip string, port string, body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in HandlePing")
		return nil
	}
	pubKeyStr, ok := bodyMap["public_key"].(string)
	fmt.Println("public_key in handle ping:", pubKeyStr)
	if !ok || pubKeyStr == "" {
		fmt.Println("Missing or invalid public_key in HandlePing")
		return nil
	}
	nodeID := node.GenerateNodeID(pubKeyStr)

	senderNode := &models.Node{
		NodeId:    nodeID,
		IP:        ip,
		Port:      port,
		PublicKey: pubKeyStr,
	}

	if GlobalPinger == nil {
		fmt.Println("❌ Pinger not registered")
		return nil
	}
	rt.InsertNode(senderNode, GlobalPinger)

	fmt.Printf("Ping from node ID: %x, IP: %s Port:%s\n", nodeID, senderNode.IP, senderNode.Port)
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

func FindNodeHandler(ip string, port string, body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in FindNodeHandler")
		return nil
	}
	pubKeyStr, ok := bodyMap["public_key"].(string)
	if !ok || pubKeyStr == "" {
		fmt.Println("Missing or invalid public_key in FindNodeHandler")
		return nil
	}
	keyStr, ok := bodyMap["node_id"].(string)
	if !ok || pubKeyStr == "" {
		fmt.Println("Missing or invalid public_key in FindNodeHandler")
		return nil
	}
	nodeID := node.GenerateNodeID(pubKeyStr)

	senderNode := &models.Node{
		NodeId:    nodeID,
		IP:        ip,
		Port:      port,
		PublicKey: pubKeyStr,
	}
	decKey, err := node.DecodeNodeID(keyStr)
	if err != nil {
		fmt.Println("Error decoding node ID:", err)
		return nil
	}

	if GlobalPinger == nil {
		fmt.Println("❌ Pinger not registered")
		return nil
	}
	rt.InsertNode(senderNode, GlobalPinger)

	closest := SendFindNode(decKey, rt)
	for _, n := range closest {
		if n.PublicKey != "" {
			n.NodeId = node.GenerateNodeID(n.PublicKey)
		}
	}
	data, err := json.Marshal(closest)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}
	return data
}

func StoreHandler(body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in StoreHandler")
		return nil
	}
	pubKeyStr, ok := bodyMap["public_key"].(string)
	if !ok || pubKeyStr == "" {
		fmt.Println("Missing or invalid public_key in StoreHandler")
		return nil
	}
	// Optionally, validate and use public_key for senderNode if needed

	var msgcert models.MsgCert
	jsonBytes, _ := json.Marshal(bodyMap)
	if err := json.Unmarshal(jsonBytes, &msgcert); err != nil {
		fmt.Println("Error unmarshaling into MsgCert:", err)
		return nil
	}

	tsmin := msgcert.Msg.Ts
	tsmin = tsmin - (tsmin % 60)
	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest := StoreValue(keyBytes, &msgcert, localNode, rt)

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
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}
	return data
}

func DeleteHandler(body interface{}, localNode *models.Node, rt *routing.RoutingTable) []byte {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in DeleteHandler")
		return nil
	}
	pubKeyStr, ok := bodyMap["public_key"].(string)
	if !ok || pubKeyStr == "" {
		fmt.Println("Missing or invalid public_key in DeleteHandler")
		return nil
	}
	// Optionally, validate and use public_key for senderNode if needed

	var repCert models.ReportCert
	jsonBytes, _ := json.Marshal(bodyMap)
	if err := json.Unmarshal(jsonBytes, &repCert); err != nil {
		fmt.Println("Error unmarshaling into ReportCert:", err)
		return nil
	}

	tsmin := repCert.Msgcert.Msg.Ts
	tsmin = tsmin - (tsmin % 60)
	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest, err := DeleteValue(&keyBytes, &repCert, localNode, rt)

	if err != nil {
		fmt.Println("Error", err)
		type ErrorResponse struct {
			Type  string `json:"type"`
			Error error  `json:"error"`
		}
		resp := ErrorResponse{
			Type:  "redirect",
			Error: err,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error while marshiling the PingResponse: ", err)
		}
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
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error while marshiling the PingResponse: ", err)
		}
		return data
	}

	fmt.Println("Store at: ", localNode)
	resp := StoredResponse{
		Type:   "deleted",
		Status: "ok",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}
	return data
}
