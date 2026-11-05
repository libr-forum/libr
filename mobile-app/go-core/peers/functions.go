package peer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
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
		log.Printf("[ERROR] Peer.Start failed: %v", err)
		return err
	}
	return nil
}

func GET(targetPeerID string, route string) ([]byte, error) {
	if Peer == nil || Peer.Host == nil {
		return nil, fmt.Errorf("libp2p peer not initialized")
	}
	reqparams := make(map[string]string)
	parts := strings.Split(route, "/")
	params := strings.Split(parts[1], "&&")

	for i := 0; i < len(params); i++ {
		key := strings.Split(params[i], "=")[0]
		value := strings.Split(params[i], "=")[1]
		reqparams[key] = value
	}
	reqparams["Method"] = "GET"
	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	GetResp, err := Peer.Send(ctx, targetPeerID, jsonReq, nil)
	if err != nil {
		fmt.Println("Error Sending get message:", err)
		return nil, err
	}
	return bytes.TrimRight(GetResp, "\x00"), nil
}

func POST(targetPeerID string, route string, body []byte) ([]byte, error) {
	if Peer == nil || Peer.Host == nil {
		return nil, fmt.Errorf("libp2p peer not initialized")
	}
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reqparams := make(map[string]string)
	parts := strings.Split(route, "/")
	params := strings.Split(parts[1], "&&")
	for i := 0; i < len(params); i++ {
		key := strings.Split(params[i], "=")[0]
		value := strings.Split(params[i], "=")[1]
		reqparams[key] = value
	}
	reqparams["Method"] = "POST"

	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		return nil, err
	}

	GetResp, err := Peer.Send(timeoutCtx, targetPeerID, jsonReq, body)
	if err != nil {
		return nil, err
	}
	return bytes.TrimRight(GetResp, "\x00"), nil
}

// MessageHandler is an interface that can be implemented in native code.
type MessageHandler interface {
	OnMessage(peerID string, msg string)
}

var (
	msgHandler MessageHandler
)

// SetMessageHandler registers a handler for incoming messages.
func SetMessageHandler(h MessageHandler) {
	msgHandler = h
}

func ServeGetReq(params []byte) []byte {
	return []byte("GET stub")
}

func ServePostReq(addr string, paramsBytes []byte, bodyBytes []byte) []byte {
	if msgHandler != nil {
		content := string(bodyBytes)
		// Try to unquote if it's a JSON string
		var unmarshaled string
		if err := json.Unmarshal(bodyBytes, &unmarshaled); err == nil {
			content = unmarshaled
		}
		msgHandler.OnMessage(addr, content)
	}
	return []byte("\"ACK\"")
}
