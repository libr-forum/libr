package node

import "math/big"

type Node struct {
	NodeId   [20]byte `json:"nodeid"`
	IP       string   `json:"ip"`
	Port     string   `json:"port"`
	LastSeen int64    `json:"lastseen"`
}

type DistanceNode struct {
	Node     *Node
	Distance *big.Int
}
