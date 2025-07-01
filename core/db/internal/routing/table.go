package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/node"
)

// Pinger interface allows us to inject ping logic from the network package.
type Pinger interface {
	Ping(selfID [20]byte, selfPort string, target node.Node) error
}

type RoutingTable struct {
	SelfID   [20]byte      `json:"self_id"`
	SelfPort string        `json:"self_port"`
	Buckets  [160]*KBucket `json:"buckets"`
}

func GetBucketIndex(selfID, targetID [20]byte) int {
	xor := node.XORBigInt(selfID, targetID)
	index := xor.BitLen() - 1
	if index < 0 {
		index = 0
	}
	return index
}

func (rt *RoutingTable) InsertNode(newNode *node.Node, pinger Pinger) string {
	index := GetBucketIndex(rt.SelfID, newNode.NodeId)

	if rt.Buckets[index] == nil {
		rt.Buckets[index] = &KBucket{}
	}
	newNode.LastSeen = time.Now().Unix()

	// ✅ Log incoming node details
	fmt.Printf("📥 InsertNode: %x | IP: %s | Port: %s\n", newNode.NodeId, newNode.IP, newNode.Port)

	return InsertNodeKBucket(rt.SelfID, rt.SelfPort, newNode, rt.Buckets[index], pinger)
}

func InsertNodeKBucket(selfID [20]byte, selfPort string, newNode *node.Node, bucket *KBucket, pinger Pinger) string {
	for i, existing := range bucket.Nodes {
		if bytes.Equal(existing.NodeId[:], newNode.NodeId[:]) {
			// ✅ Update existing node info including IP/Port/LastSeen
			existing.IP = newNode.IP
			existing.Port = newNode.Port
			existing.LastSeen = newNode.LastSeen

			bucket.Nodes = append(bucket.Nodes[:i], bucket.Nodes[i+1:]...)
			bucket.Nodes = append(bucket.Nodes, existing)

			fmt.Printf("🔁 Updated node in K-bucket: %x | Port: %s\n", newNode.NodeId, newNode.Port)
			return "Updated K-Bucket (refreshed existing node)"
		}
	}

	if len(bucket.Nodes) < config.K {
		bucket.Nodes = append(bucket.Nodes, newNode)
		fmt.Printf("➕ Appended new node: %x | Port: %s\n", newNode.NodeId, newNode.Port)
		return "Appended new node (bucket had space)"
	}

	// Ping the oldest node to check if it’s alive
	if err := pinger.Ping(selfID, selfPort, *bucket.Nodes[0]); err != nil {
		fmt.Printf("⚠️ Oldest node unresponsive. Replacing with: %x | Port: %s\n", newNode.NodeId, newNode.Port)
		bucket.Nodes = append(bucket.Nodes[1:], newNode)
		return "Replaced unresponsive node with new node"
	}

	fmt.Println("🚫 New node rejected (bucket full, oldest still active)")
	return "New node rejected (bucket full, oldest still active)"
}

func (rt *RoutingTable) FindClosest(targetID [20]byte, count int) []*node.Node {
	var allNodes []*node.Node

	for _, bucket := range rt.Buckets {
		if bucket == nil {
			continue
		}
		allNodes = append(allNodes, bucket.Nodes...)
	}

	sort.Slice(allNodes, func(i, j int) bool {
		distI := node.XORBigInt(allNodes[i].NodeId, targetID)
		distJ := node.XORBigInt(allNodes[j].NodeId, targetID)
		return distI.Cmp(distJ) < 0
	})

	if len(allNodes) > count {
		return allNodes[:count]
	}
	return allNodes
}

func (rt *RoutingTable) SaveTo(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(rt)
}

func LoadRoutingTable(filename string) (*RoutingTable, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rt RoutingTable
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func NewRoutingTable(selfID [20]byte, selfPort string) *RoutingTable {
	rt := &RoutingTable{
		SelfID:   selfID,
		SelfPort: selfPort,
	}

	for i := range rt.Buckets {
		rt.Buckets[i] = &KBucket{}
	}

	return rt
}

func (rt *RoutingTable) SelfIDHex() string {
	return fmt.Sprintf("%x", rt.SelfID[:])
}

var memoryCache *RoutingTable

func GetOrCreateRoutingTable(node *node.Node) *RoutingTable {
	if memoryCache != nil {
		return memoryCache
	}

	// Try to load from DB
	dbRT, err := LoadRoutingTableFromDB()
	if err == nil {
		memoryCache = dbRT
		return memoryCache
	}

	// If failed to load from DB, create a new one
	memoryCache = NewRoutingTable(node.NodeId, node.Port)
	go memoryCache.SaveToDBAsync() // save it asynchronously
	return memoryCache
}

func (rt *RoutingTable) SaveToDBAsync() {
	go func() {
		jsonBytes, err := json.Marshal(rt)
		if err != nil {
			fmt.Println("Error marshaling routing table:", err)
			return
		}

		_, err = config.Pool.Exec(context.Background(), `
			INSERT INTO RoutingTable (rt) VALUES ($1)
		`, jsonBytes)
		if err != nil {
			fmt.Println("Error saving routing table to DB:", err)
		}
	}()
}

func LoadRoutingTableFromDB() (*RoutingTable, error) {
	var jsonBytes []byte
	err := config.Pool.QueryRow(context.Background(), `
		SELECT rt FROM RoutingTable ORDER BY id DESC LIMIT 1
	`).Scan(&jsonBytes)
	if err != nil {
		return nil, err
	}

	var rt RoutingTable
	err = json.Unmarshal(jsonBytes, &rt)
	return &rt, err
}
