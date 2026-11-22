package peer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

var Peer *ChatPeer

type RelayDist struct {
	relayID string
	dist    *big.Int
}

func StartNode(relayMultiAddrList []string, onStatusChange func(string), opts ...libp2p.Option) error {

	fmt.Println("Starting Node...")
	var err error
	Peer, err = NewChatPeer(relayMultiAddrList, opts...)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return err
	}

	ctx := context.Background()

	Peer.OnStatusChange = onStatusChange
	if err := Peer.Start(ctx); err != nil {
		log.Printf("[ERROR] Peer.Start failed: %v", err)
		return err
	}
	return nil
}

func GET(targetPeerID string, route string) ([]byte, error) { //"/ts=123&&id=123"
	if Peer == nil || Peer.Host == nil {
		return nil, errors.New("chat peer is not initialized")
	}

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
	if Peer == nil || Peer.Host == nil {
		return nil, errors.New("chat peer is not initialized")
	}

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

// PostHandler serves one POST route; from is the caller's address, body the raw JSON body.
type PostHandler func(from string, body []byte) []byte

var (
	postMu       sync.RWMutex
	postRoutes   = map[string]PostHandler{}
	postFallback PostHandler
)

// HandlePost registers h for route. Moderators register "auto" and "manual"
// (mod/handlers); peers that cannot moderate register nothing.
func HandlePost(route string, h PostHandler) {
	postMu.Lock()
	postRoutes[route] = h
	postMu.Unlock()
}

// HandleUnknownPost registers h for POSTs on routes nobody registered (the mobile
// bridge forwards them to native code).
func HandleUnknownPost(h PostHandler) {
	postMu.Lock()
	postFallback = h
	postMu.Unlock()
}

func ServePostReq(addr string, paramsBytes []byte, bodyBytes []byte) []byte {
	var params map[string]interface{}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		fmt.Println("Failed to unmarshal params:", err)
		return nil
	}
	route, _ := params["route"].(string)

	postMu.RLock()
	h, ok := postRoutes[route]
	if !ok {
		h = postFallback
	}
	postMu.RUnlock()
	if h == nil {
		fmt.Println("Unknown POST route:", route)
		return nil
	}
	resp := h(addr, bodyBytes)
	if resp == nil {
		fmt.Printf("[ERROR] POST route %q returned nil\n", route)
	}
	return resp
}
