package routing

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/node"
)

// Pinger interface allows us to inject ping logic from the network package.
type Pinger interface {
	Ping(selfID [20]byte, target node.Node) error
}

type RoutingTable struct {
	SelfID  [20]byte
	Buckets [160]*KBucket
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

	return InsertNodeKBucket(rt.SelfID, newNode, rt.Buckets[index], pinger)
}

func InsertNodeKBucket(selfID [20]byte, newNode *node.Node, bucket *KBucket, pinger Pinger) string {
	for i, existing := range bucket.Nodes {
		if bytes.Equal(existing.NodeId[:], newNode.NodeId[:]) {
			bucket.Nodes = append(bucket.Nodes[:i], bucket.Nodes[i+1:]...)
			bucket.Nodes = append(bucket.Nodes, newNode)
			return "Updated K-Bucket (refreshed existing node)"
		}
	}

	if len(bucket.Nodes) < config.K {
		bucket.Nodes = append(bucket.Nodes, newNode)
		return "Appended new node (bucket had space)"
	}

	if err := pinger.Ping(selfID, *bucket.Nodes[0]); err != nil {
		bucket.Nodes = append(bucket.Nodes[1:], newNode)
		return "Replaced unresponsive node with new node"
	}

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

func NewRoutingTable(selfID [20]byte) *RoutingTable {
	rt := &RoutingTable{
		SelfID: selfID,
	}

	for i := range rt.Buckets {
		rt.Buckets[i] = &KBucket{}
	}

	return rt
}
