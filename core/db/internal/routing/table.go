package routing

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/node"
)

var GlobalRT *RoutingTable

// Pinger interface allows us to inject ping logic from the network package.
type Pinger interface {
	Ping(peerId string, target *models.Node) error
}

type RoutingTable struct {
	SelfID  [20]byte             `json:"self_id"`
	Buckets [160]*models.KBucket `json:"buckets"`
}

func GetBucketIndex(selfID, targetID [20]byte) int {
	xor := node.XORBigInt(selfID, targetID)
	index := xor.BitLen() - 1
	if index < 0 {
		index = 0
	}
	return index
}

func (rt *RoutingTable) InsertNode(localNode *models.Node, newNode *models.Node, pinger Pinger) string {
	if bytes.Equal(rt.SelfID[:], newNode.NodeId[:]) {
		return "Can't add self node"
	}

	index := GetBucketIndex(rt.SelfID, newNode.NodeId)
	newNode.BucketIdx = index

	if rt.Buckets[index] == nil {
		rt.Buckets[index] = &models.KBucket{}
	}
	newNode.LastSeen = time.Now().Unix()

	// ✅ Log incoming node details
	fmt.Printf("📥 InsertNode: %x | PeerID: %s\n", newNode.NodeId, newNode.PeerId)
	return InsertNodeKBucket(rt.SelfID, localNode, newNode, rt.Buckets[index], pinger)
}

func InsertNodeKBucket(selfID [20]byte, localNode *models.Node, newNode *models.Node, bucket *models.KBucket, pinger Pinger) string {
	for i, existing := range bucket.Nodes {
		// ✅ Update existing node info including PeerID/LastSeen
		if bytes.Equal(existing.NodeId[:], newNode.NodeId[:]) {
			existing.LastSeen = newNode.LastSeen

			bucket.Nodes = append(bucket.Nodes[:i], bucket.Nodes[i+1:]...)
			bucket.Nodes = append(bucket.Nodes, existing)

			fmt.Printf("🔁 Updated node in K-bucket: %x | Port: %s\n", newNode.NodeId, newNode.PeerId)
			return "Updated K-Bucket (refreshed existing node)"
		}
	}

	if len(bucket.Nodes) < config.K {
		bucket.Nodes = append(bucket.Nodes, newNode)
		fmt.Printf("➕ Appended new node: %x | Port: %s\n", newNode.NodeId, newNode.PeerId)
		return "Appended new node (bucket had space)"
	}

	// Ping the oldest node to check if it’s alive
	if err := pinger.Ping(localNode.PeerId, bucket.Nodes[0]); err != nil {
		fmt.Printf("⚠️ Oldest node unresponsive. Replacing with: %x | Port: %s\n", newNode.NodeId, newNode.PeerId)
		bucket.Nodes = append(bucket.Nodes[1:], newNode)
		return "Replaced unresponsive node with new node"
	}

	fmt.Println("🚫 New node rejected (bucket full, oldest still active)")

	return "New node rejected (bucket full, oldest still active)"
}

func (rt *RoutingTable) FindClosest(targetID [20]byte, count int) []*models.Node {
	var allNodes []*models.Node
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

func NewRoutingTable(selfID [20]byte) *RoutingTable {
	rt := &RoutingTable{
		SelfID: selfID,
	}
	for i := range rt.Buckets {
		rt.Buckets[i] = &models.KBucket{}
	}
	return rt
}

func (rt *RoutingTable) SelfIDHex() string {
	return fmt.Sprintf("%x", rt.SelfID[:])
}

var memoryCache *RoutingTable

func GetOrCreateRoutingTable(selfID [20]byte) *RoutingTable {
	if memoryCache != nil {
		return memoryCache
	}

	dbRT, err := LoadRoutingTable(selfID)
	if err == nil {
		memoryCache = dbRT
		return memoryCache
	}

	memoryCache = NewRoutingTable(selfID)
	go memoryCache.SaveToDBAsync()
	return memoryCache
}

// func (rt *RoutingTable) SaveToDBAsync() {
// 	go func() {
// 		jsonBytes, err := json.Marshal(rt)
// 		if err != nil {
// 			fmt.Println("❌ Error marshaling routing table:", err)
// 			return
// 		}

// 		_, err = config.DB.ExecContext(context.Background(),
// 			`INSERT INTO RoutingTable (rt) VALUES (?)`, jsonBytes)
// 		if err != nil {
// 			fmt.Println("❌ Error saving routing table to SQLite:", err)
// 		}
// 	}()
// }

// func LoadRoutingTableFromDB() (*RoutingTable, error) {
// 	var jsonBytes []byte
// 	row := config.DB.QueryRowContext(context.Background(),
// 		`SELECT rt FROM RoutingTable ORDER BY id DESC LIMIT 1`)
// 	err := row.Scan(&jsonBytes)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("no routing table in DB")
// 		}
// 		return nil, err
// 	}

// 	var rt RoutingTable
// 	if err := json.Unmarshal(jsonBytes, &rt); err != nil {
// 		return nil, fmt.Errorf("error unmarshaling routing table: %v", err)
// 	}
// 	return &rt, nil
// }

func LoadRoutingTable(selfID [20]byte) (*RoutingTable, error) {
	fmt.Println("[DEBUG] LoadRoutingTable called")

	rt := &RoutingTable{
		SelfID: selfID,
	}
	fmt.Println("[DEBUG] Created empty RoutingTable with SelfID:", selfID)

	// Initialize empty buckets
	for i := 0; i < len(rt.Buckets); i++ {
		rt.Buckets[i] = &models.KBucket{Nodes: []*models.Node{}}
	}
	fmt.Println("[DEBUG] Initialized", len(rt.Buckets), "empty buckets")

	query := `
		SELECT bucket_idx, NodeID, PeerID, LastSeen
		FROM RoutingTable
		ORDER BY bucket_idx ASC
	`
	fmt.Println("[DEBUG] Executing query:", query)

	rows, err := config.DB.Query(query)
	if err != nil {
		fmt.Println("[ERROR] DB.Query failed:", err)
		return nil, err
	}
	defer func() {
		fmt.Println("[DEBUG] Closing DB rows")
		rows.Close()
	}()

	rowCount := 0
	for rows.Next() {
		fmt.Println("[DEBUG] Processing next row...")
		var (
			bucketIdx int
			nodeIDRaw []byte
			peerID    string
			lastSeen  int64
		)

		if err := rows.Scan(&bucketIdx, &nodeIDRaw, &peerID, &lastSeen); err != nil {
			fmt.Println("[ERROR] rows.Scan failed:", err)
			return nil, err
		}

		fmt.Printf("[DEBUG] Scanned row #%d - bucketIdx: %d, nodeIDRaw(len=%d): %x, peerID: %s, lastSeen: %d\n",
			rowCount, bucketIdx, len(nodeIDRaw), nodeIDRaw, peerID, lastSeen)

		if bucketIdx < 0 || bucketIdx >= len(rt.Buckets) {
			fmt.Printf("[WARN] Invalid bucket index %d, skipping this row\n", bucketIdx)
			continue
		}

		if len(nodeIDRaw) != 20 {
			fmt.Printf("[WARN] NodeID length is %d (expected 20), skipping\n", len(nodeIDRaw))
			continue
		}

		var nodeID [20]byte
		copy(nodeID[:], nodeIDRaw)

		node := &models.Node{
			NodeId:    nodeID,
			PeerId:    peerID,
			BucketIdx: bucketIdx,
			LastSeen:  lastSeen,
		}

		rt.Buckets[bucketIdx].Nodes = append(rt.Buckets[bucketIdx].Nodes, node)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		fmt.Println("[ERROR] rows.Err() returned:", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] Loaded %d nodes into routing table\n", rowCount)
	return rt, nil
}

func (rt *RoutingTable) SaveToDBAsync() {
	go func() {
		ctx := context.Background()

		tx, err := config.DB.BeginTx(ctx, nil)
		if err != nil {
			fmt.Println("❌ Error starting transaction:", err)
			return
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT OR REPLACE INTO RoutingTable (bucket_idx, NodeID, PeerID, LastSeen)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			fmt.Println("❌ Error preparing statement:", err)
			_ = tx.Rollback()
			return
		}
		defer stmt.Close()

		for bucketIdx, bucket := range rt.Buckets {
			if bucket == nil {
				continue
			}
			for _, node := range bucket.Nodes {
				_, err := stmt.ExecContext(
					ctx,
					bucketIdx,
					node.NodeId[:],
					node.PeerId,
					node.LastSeen,
				)
				if err != nil {
					fmt.Println("❌ Error inserting node:", err)
					_ = tx.Rollback()
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			fmt.Println("❌ Error committing transaction:", err)
		}
	}()
}
