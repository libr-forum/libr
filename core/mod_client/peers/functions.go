package peer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/mod_client/internal/handlers"
)

var Peer *ChatPeer

type RelayDist struct {
	relayID string
	dist    *big.Int
}

func StartNode(relayMultiAddrList []string) error {

	fmt.Println("Starting Node...")
	var err error
	Peer, err = NewChatPeer(relayMultiAddrList)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return err
	}

	ctx := context.Background()

	if err := Peer.Start(ctx); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
	// initDHT()
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

	GetResp, err := Peer.Send(timeoutCtx, targetPeerID, jsonReq, body)

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

	GetResp = bytes.TrimRight(GetResp, "\x00")
	return GetResp, nil
}

func ServeGetReq([]byte) []byte {

	//add logic to serve get requests here

	var resp []byte
	return resp

}

func ServePostReq(addr string, paramsBytes []byte, bodyBytes []byte) []byte {
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

	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		fmt.Println("Failed to unmarshal body:", err)
		return nil
	}
	fmt.Println("Body:", body)
	switch route {
	case "auto":
		return handlers.MsgIN(bodyBytes)

	case "manual":
		return handlers.MsgReport(bodyBytes)

	default:
		fmt.Println("Unknown POST route:", route)
		return nil
	}
}
