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

func HandlePing(ip string, body interface{}, localNode *node.Node, rt *routing.RoutingTable) []byte {
	var pingReq PingRequest

	// Unmarshal into pingReq
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in HandlePing")
		return nil
	}
	if nodeIDStr, ok := bodyMap["node_id"].(string); ok {
		pingReq.NodeID = nodeIDStr
	}

	dedID, err := node.DecodeNodeID(pingReq.NodeID)
	if err != nil {
		fmt.Println("Failed to decode NodeID")
		return nil
	}

	senderNode := &node.Node{
		NodeId: dedID,
		IP:     ip,
		Port:   bodyMap["port"].(string),
	}

	if GlobalPinger == nil {
		fmt.Println("❌ Pinger not registered")
		return nil
	}
	rt.InsertNode(senderNode, GlobalPinger)

	fmt.Printf("Ping from node ID: %x, IP: %s Port:%s\n", dedID, senderNode.IP, senderNode.Port)
	data, err := json.Marshal(PingResponse{Status: "ok"})
	if err != nil {
		fmt.Println("Error while marshaling the PingResponse: ", err)
	}

	return data
}

func FindValueHandler(key string, localNode *node.Node, rt *routing.RoutingTable) []byte {

	values, closest := SendFindValue(key, localNode, rt)

	if values != nil {
		fmt.Println("Found the value")
		type FoundResponse struct {
			Type   string           `json:"type"`
			Values []models.MsgCert `json:"values"`
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
			Type  string       `json:"type"`
			Nodes []*node.Node `json:"nodes"`
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

func FindNodeHandler(key [20]byte, localNode *node.Node, rt *routing.RoutingTable) []byte {
	closest := SendFindNode(key, rt)

	data, err := json.Marshal(closest)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}

	return data
}

func StoreHandler(body models.MsgCert, localNode *node.Node, rt *routing.RoutingTable) []byte {

	msgcert := body
	fmt.Println(msgcert)

	tsmin := msgcert.Msg.Ts
	tsmin = tsmin - (tsmin % 60)
	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest := StoreValue(keyBytes, msgcert, localNode, rt)

	if closest != nil {
		fmt.Println("Sending list of k closest nodes")
		type RedirectResponse struct {
			Type  string       `json:"type"`
			Nodes []*node.Node `json:"nodes"`
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
