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
	Port   string `json:"port"`
}

type PingResponse struct {
	Status string `json:"status"`
}

type StoredResponse struct {
	Type   string `json:"type"`
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

		fmt.Println("Ping Req: ", pingReq)
		senderNode := &node.Node{
			NodeId: dedID,
			IP:     r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")],
			Port:   pingReq.Port,
		}

		pinger := &network.RealPinger{}
		rt.InsertNode(senderNode, pinger)

		json.NewEncoder(w).Encode(PingResponse{Status: "ok"})
		fmt.Printf("Ping from node ID: %x, IP: %s Port:%s\n", dedID, senderNode.IP, senderNode.Port)

		fmt.Println("🔍 Routing Table Dump:")
		for i, bucket := range rt.Buckets {
			if bucket == nil || len(bucket.Nodes) == 0 {
				continue
			}
			fmt.Printf("Bucket %2d:\n", i)
			for _, n := range bucket.Nodes {
				fmt.Printf("  - ID: %x | IP: %s | Port: %s | LastSeen: %d\n",
					n.NodeId, n.IP, n.Port, n.LastSeen)
			}
		}

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
			type FoundResponse struct {
				Type   string           `json:"type"`
				Values []models.MsgCert `json:"values"`
			}
			resp := FoundResponse{
				Type:   "found",
				Values: values,
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			type RedirectResponse struct {
				Type  string       `json:"type"`
				Nodes []*node.Node `json:"nodes"`
			}
			resp := RedirectResponse{
				Type:  "redirect",
				Nodes: closest,
			}
			json.NewEncoder(w).Encode(resp)
		}
	}
}

func StoreHandler(localNode *node.Node, rt *routing.RoutingTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var Msgcert models.MsgCert
		if err := json.NewDecoder(r.Body).Decode(&Msgcert); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		key := strconv.FormatInt(Msgcert.Msg.Ts, 10)
		keyBytes := node.GenerateNodeID(key)

		closest := network.StoreValue(keyBytes, Msgcert, localNode, rt)

		w.Header().Set("Content-Type", "application/json")

		if closest != nil {
			type RedirectResponse struct {
				Type  string       `json:"type"`
				Nodes []*node.Node `json:"nodes"`
			}
			resp := RedirectResponse{
				Type:  "redirect",
				Nodes: closest,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := StoredResponse{
			Type:   "stored",
			Status: "ok",
		}
		json.NewEncoder(w).Encode(resp)
	}
}
