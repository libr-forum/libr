package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const K = 4

var Pool *pgxpool.Pool

func EnsureDatabaseExists(uri string) {
	fmt.Println("Trying to connect to db")
	dbName := "libr"
	ctx := context.Background()
	var exists bool

	// First connect to the default "postgres" DB to check/create "libr"
	newURI := fmt.Sprintf(
		"postgres://%s:%s@postgres:%s/postgres?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_PORT"),
	)

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

	// Reconnect to the newly created (or existing) "libr" database
	Pool, err = pgxpool.New(ctx, uri)
	if err != nil {
		log.Fatalf("Unable to connect to 'libr' database: %v", err)
	}

	// Create MsgCert table
	createMsgCertSQL := `
	CREATE TABLE IF NOT EXISTS msgcert (
		sender TEXT NOT NULL,
		content TEXT NOT NULL,
		ts TIMESTAMPTZ NOT NULL,
		mod_certs JSONB NOT NULL,
		sign TEXT NOT NULL
	)`
	if _, err := Pool.Exec(ctx, createMsgCertSQL); err != nil {
		log.Fatalf("Failed to create msgcert table: %v", err)
	}

	// Create RoutingTable table
	createRoutingSQL := `
	CREATE TABLE IF NOT EXISTS RoutingTable (
		id SERIAL PRIMARY KEY,
		rt JSONB NOT NULL
	)`
	if _, err := Pool.Exec(ctx, createRoutingSQL); err != nil {
		log.Fatalf("Failed to create RoutingTable table: %v", err)
	}
}

func InitConnection(port string) {
	// Only load env file here if not already loaded
	envPath := "./core/db/.env"
	if port != "" {
		envPath = fmt.Sprintf("./core/db/.env.%s", port)
	}

	// Load only if not already loaded
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("warning: failed to load %s: %v", envPath, err)
		// Do NOT fatal — env might already be loaded by initDHT
	}

	// Read DB params from env
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	if dbUser == "" || dbPass == "" || dbPort == "" {
		log.Fatal("DB_USER, DB_PASS, or DB_PORT not set in environment")
	}

	// Use "libr" as hardcoded DB name for now
	uri := fmt.Sprintf(
		"postgres://%s:%s@postgres:%s/libr?sslmode=disable",
		dbUser, dbPass, dbPort,
	)

	EnsureDatabaseExists(uri)

	var err error
	Pool, err = pgxpool.New(context.Background(), uri)
	if err != nil {
		log.Fatal("failed to create DB pool:", err)
	}

	log.Println("connected to db")
}
