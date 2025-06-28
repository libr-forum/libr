package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

func EnsureDatabaseExists(uri string) {
	fmt.Println("Trying to connect to db")
	dbName := "libr"
	ctx := context.Background()
	var exists bool

	newURI := fmt.Sprintf("postgres://%s:%s@localhost:5432/postgres?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"))

	var err error
	Pool, err = pgxpool.New(ctx, newURI)
	if err != nil {
		log.Fatal("couldn't connect to postgres:", err)
	}

	err = Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM pg_catalog.pg_database
			WHERE datname = $1
		)`, dbName).Scan(&exists)
	if err != nil {
		log.Fatalf("checking of libr failed: %v", err)
	}

	if !exists {
		log.Printf("Database %q not found – creating...", dbName)
		if _, err := Pool.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)); err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("Database %q created.", dbName)
	} else {
		log.Printf("Database %q already exists.", dbName)
	}

	// Connect to the new 'libr' database
	Pool, err = pgxpool.New(ctx, uri)
	if err != nil {
		log.Fatalf("Unable to connect to 'libr' database: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS MsgCert (
		sender TEXT NOT NULL,
		content TEXT NOT NULL,
		ts TIMESTAMPTZ NOT NULL,
		mod_cert JSONB NOT NULL,
		sign TEXT NOT NULL
	)`
	if _, err := Pool.Exec(ctx, createTableSQL); err != nil {
		log.Fatalf("Failed to create MsgCert table: %v", err)
	}
}

func InitConnection() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file:", err)
	}
	uri := fmt.Sprintf(
		"postgres://%s:%s@localhost:5432/libr?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
	)
	EnsureDatabaseExists(uri)

	var err error
	Pool, err = pgxpool.New(context.Background(), uri)
	if err != nil {
		log.Fatal("failed to create pool:", err)
	}
	log.Println("connected to db")

}

// --> timestamp: check and set as int

// func main() {
// 	InitConnection()
// 	    sample := models.MsgCert{
//         Sender:    "sender_public_key_example",
//         Msg:       "Hello, Libr!",
//         Timestamp: "1711578607",
//         ModCert: []models.ModCert{{
//             PublicKey: "mod_public_key_example",
//             Sign:      "signature_example",
//             Status:    "approved",
//         }},
//     }

//     if msg, err := InsertMsgCert(sample); err != nil {
//         log.Printf("Insert error: %v", err)
//     } else {
//         log.Println(msg)
//     }

// 	defer Pool.Close()

// }
