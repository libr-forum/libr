package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/network"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

type PingRequest struct {
	NodeID string `json:"node_id"`
}

type PingResponse struct {
	Status string `json:"status"`
}

func HandlePing(localNode *node.Node, rt *routing.RoutingTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received /ping from:", r.RemoteAddr)

		var pingReq PingRequest
		if err := json.NewDecoder(r.Body).Decode(&pingReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		dedID, err := node.DecodeNodeID(pingReq.NodeID)
		if err != nil {
			http.Error(w, "Invalid node ID format", http.StatusBadRequest)
			return
		}

		senderNode := &node.Node{
			NodeId: dedID,
			IP:     r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")],
			Port:   "",
		}

		pinger := &network.RealPinger{}
		rt.InsertNode(senderNode, pinger)

		json.NewEncoder(w).Encode(PingResponse{Status: "ok"})
	}
}

func FindNodeHandler(localNode *node.Node, rt *routing.RoutingTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("id")
		if query == "" {
			http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
			return
		}

		targetID, err := node.DecodeNodeID(query)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		closest := network.SendFindNode(targetID, rt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(closest)
	}
}

func FindValueHandler(localNode *node.Node, rt *routing.RoutingTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		values, closest := network.SendFindValue(key, localNode, rt)
		w.Header().Set("Content-Type", "application/json")

		if values != nil {
			json.NewEncoder(w).Encode(values)
		} else {
			json.NewEncoder(w).Encode(closest)
		}
	}
}

func StoreHandler(localNode *node.Node, rt *routing.RoutingTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var cert models.MsgCert
		if err := json.NewDecoder(r.Body).Decode(&cert); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		key := strconv.FormatInt(cert.Msg.Ts, 10)
		keyBytes := node.GenerateNodeID(key)

		closest := network.StoreValue(keyBytes, cert, localNode, rt)
		if closest != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(closest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Stored successfully on k-closest node")
	}
}
