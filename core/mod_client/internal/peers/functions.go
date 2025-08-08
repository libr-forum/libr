package lmao

// import (
// 	"bytes"
// 	"context"

// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"math/big"
// 	"strings"
// 	"time"

// 	"github.com/devlup-labs/Libr/core/mod_client/internal/handlers"
// )

// var Peer *ChatPeer

// type RelayDist struct {
// 	relayID string
// 	dist    *big.Int
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

// func ServeGetReq([]byte) []byte {

// 	//add logic to serve get requests here

// 	var resp []byte
// 	return resp

// }

// func ServePostReq(paramsBytes []byte, bodyBytes []byte) []byte {
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
// 	fmt.Println(route, string(bodyBytes))
// 	switch route {
// 	case "auto":
// 		return handlers.MsgIN(bodyBytes)

// 	case "manual":
// 		return handlers.MsgReport(bodyBytes)

// 	default:
// 		fmt.Println("Unknown POST route:", route)
// 		return nil
// 	}
// }

// func XorHexToBigInt(hex1, hex2 string) *big.Int {

// 	bytes1, err1 := hex.DecodeString(hex1)
// 	bytes2, err2 := hex.DecodeString(hex2)

// 	if err1 != nil || err2 != nil {
// 		log.Fatalf("Error decoding hex: %v %v", err1, err2)
// 	}

// 	if len(bytes1) != len(bytes2) {
// 		log.Fatalf("Hex strings must be the same length")
// 	}

// 	xorBytes := make([]byte, len(bytes1))
// 	for i := 0; i < len(bytes1); i++ {
// 		xorBytes[i] = bytes1[i] ^ bytes2[i]
// 	}

// 	result := new(big.Int).SetBytes(xorBytes)
// 	return result
// }
