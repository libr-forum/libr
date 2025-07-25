package Peers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var Peer *ChatPeer

func StartNode(relayAdd string) error {

	fmt.Println("Starting Node...")

	relayAddr := relayAdd // have to build logic about getting Multiple Relays

	var err error
	Peer, err = NewChatPeer(relayAddr)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return err
	}
	ctx := context.Background()

	if err := Peer.Start(ctx); err != nil {
		return err
	}
	return nil
}

func GET(targetIP string, targetPort string, route string) ([]byte, error) {

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
		//fmt.Println("Error Sending trial get message")
	}

	return GetResp, nil //this will be of json bytes with a field type, and other field(Resp) having the response bytes from the server
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
		//fmt.Println("Error Sending trial get message")
	}

	return GetResp, nil
}

func ServeGetReq([]byte) []byte {

	//add logic to serve get requests here

	var resp []byte
	return resp

}

func ServePostReq([]byte, []byte) []byte {
	var PostResp []byte
	return PostResp
}
