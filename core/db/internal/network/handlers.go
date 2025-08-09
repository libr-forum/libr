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
	var pingReq PingRequest

	// Unmarshal into pingReq
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid body format in HandlePing")
		return nil
	}
	var publicKeyStr string
	if pk, ok := bodyMap["public_key"].(string); ok {
		publicKeyStr = pk
	} else {
		fmt.Println("Missing public_key in ping body")
		return nil
	}
	// Always generate NodeId from public_key
	dedID := node.GenerateNodeID(publicKeyStr)

	senderNode := &models.Node{
		NodeId:    dedID,
		IP:        ip,
		Port:      port,
		PublicKey: publicKeyStr,
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

func FindNodeHandler(key [20]byte, localNode *models.Node, rt *routing.RoutingTable) []byte {
	closest := SendFindNode(key, rt)

	data, err := json.Marshal(closest)
	if err != nil {
		fmt.Println("Error while marshiling the PingResponse: ", err)
	}

	return data
}

func StoreHandler(body models.MsgCert, localNode *models.Node, rt *routing.RoutingTable) []byte {

	msgcert := body
	fmt.Println(msgcert)

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

func DeleteHandler(body *models.ReportCert, localNode *models.Node, rt *routing.RoutingTable) []byte {

	repCert := body

	tsmin := repCert.Msgcert.Msg.Ts
	tsmin = tsmin - (tsmin % 60)
	keyBytes := node.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	fmt.Println(tsmin, keyBytes)

	closest, err := DeleteValue(&keyBytes, repCert, localNode, rt) //for now err is ignored

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
