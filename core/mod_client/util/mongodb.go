package util

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/mod_client/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// SetupMongo initializes the global MongoClient
func SetupMongo(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	MongoClient = client
	log.Println("✅ MongoDB connected successfully")
	return nil
}

// DisconnectMongo gracefully closes the MongoDB connection
func DisconnectMongo() {
	if MongoClient != nil {
		if err := MongoClient.Disconnect(context.Background()); err != nil {
			log.Println("⚠️ Error disconnecting MongoDB:", err)
		} else {
			log.Println("🛑 MongoDB disconnected")
		}
	}
}

// GetStartNodes fetches all known nodes from the DB
func GetStartNodes() ([]*types.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoClient.Database("Addrs").Collection("nodes")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}
	defer cursor.Close(ctx)

	var nodeList []*types.Node
	for cursor.Next(ctx) {
		var doc struct {
			NodeId string `bson:"node_id"`
			PeerId string `bson:"peer_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		nodeId, _ := DecodeNodeID(doc.NodeId)
		node := &types.Node{
			NodeId: nodeId,
			PeerId: doc.PeerId,
		}
		nodeList = append(nodeList, node)
	}

	return nodeList, nil
}

// GetOnlineMods fetches all currently online moderators from the DB
func GetOnlineMods() ([]types.Mod, error) {
	fmt.Println("Fetching online mods from MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoClient.Database("Addrs").Collection("mods")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Error fetching mods:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var mods []types.Mod
	for cursor.Next(ctx) {
		var doc struct {
			PeerId    string `bson:"peer_id"`
			PublicKey string `bson:"public_key"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode mod document: %w", err)
		}
		mods = append(mods, types.Mod{
			PeerId:    doc.PeerId,
			PublicKey: doc.PublicKey,
		})
	}

	fmt.Println("✅ Mods fetched successfully")
	return mods, nil
}

// GetRelayAddr fetches available relay multiaddresses from the DB
func GetRelayAddr() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoClient.Database("Addrs").Collection("relays")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relay addresses: %w", err)
	}
	defer cursor.Close(ctx)

	var relayList []string
	for cursor.Next(ctx) {
		var doc struct {
			Address string `bson:"address"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode relay document: %w", err)
		}
		if strings.HasPrefix(doc.Address, "/") {
			relayList = append(relayList, strings.TrimSpace(doc.Address))
		}
	}

	return relayList, nil
}
