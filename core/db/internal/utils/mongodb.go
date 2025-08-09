package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	ctx         context.Context
	cancel      context.CancelFunc
)

// SetupMongo initializes the global MongoClient
func SetupMongo(uri string) error {
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		cancel()
		return err
	}

	// Check connection
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		return err
	}

	MongoClient = client
	log.Println("✅ MongoDB connected")
	return nil
}

// DisconnectMongo gracefully closes the MongoDB connection
func DisconnectMongo() {
	if cancel != nil {
		cancel()
	}
	if MongoClient != nil {
		if err := MongoClient.Disconnect(context.Background()); err != nil {
			log.Println("⚠️ Error disconnecting MongoDB:", err)
		} else {
			log.Println("🛑 MongoDB disconnected")
		}
	}
}

// 🚀 Uses global MongoClient and ctx
func GetDbAddr() ([]*models.Node, error) {
	collection := MongoClient.Database("Addrs").Collection("nodes") // replace with actual DB & collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var nodeList []*models.Node
	for cursor.Next(ctx) {
		var doc struct {
			IP        string `bson:"ip"`
			Port      string `bson:"port"`
			PublicKey string `bson:"public_key"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		node := &models.Node{
			NodeId:    node.GenerateNodeID(doc.PublicKey),
			IP:        doc.IP,
			Port:      doc.Port,
			PublicKey: doc.PublicKey,
		}
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

func GetOnlineMods() ([]*models.Mod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoClient.Database("Addrs").Collection("mods")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mods []*models.Mod
	for cursor.Next(ctx) {
		var doc struct {
			IP        string `bson:"ip"`
			Port      string `bson:"port"`
			PublicKey string `bson:"public_key"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		mods = append(mods, &models.Mod{
			IP:        doc.IP,
			Port:      doc.Port,
			PublicKey: doc.PublicKey,
		})
	}
	fmt.Println("Online mods:", mods)
	return mods, nil
}

func GetRelayAddr() ([]string, error) {
	collection := MongoClient.Database("Addrs").Collection("relays") // replace with actual DB & collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var relayList []string
	for cursor.Next(ctx) {
		var doc struct {
			Address string `bson:"address"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		if strings.HasPrefix(doc.Address, "/") {
			relayList = append(relayList, strings.TrimSpace(doc.Address))
		}
	}
	return relayList, nil
}
