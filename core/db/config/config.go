package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const K = 4

// DB is the global SQLite connection
var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or couldn't load it")
	}

	dbPath := getDBPath()
	err = os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create DB directory: %v", err)
	}

	DB, err = OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite DB: %v", err)
	}

	log.Println("SQLite database initialized successfully.")
}

// OpenDB opens the SQLite file at path and creates the tables. Tests point it at a temp file.
func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func getDBPath() string {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("AppData")
	case "darwin":
		baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default: // assume Linux/Unix
		baseDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	if baseDir == "" {
		return ""
	}

	dbPath := filepath.Join(baseDir, "libr", "db", "libr.db")
	return dbPath
}

func createTables(DB *sql.DB) error {
	createMsgCertTable := `
	CREATE TABLE IF NOT EXISTS msgcert (
		sender TEXT NOT NULL,
		content TEXT NOT NULL,
		ts INTEGER NOT NULL,
		mod_certs TEXT NOT NULL,
		sign TEXT NOT NULL,
		deleted INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS indx_ts ON msgcert(ts);
	CREATE INDEX IF NOT EXISTS indx_ts_sender ON msgcert(ts, sender);`

	createRoutingTable := `
	CREATE TABLE IF NOT EXISTS RoutingTable (
		bucket_idx INTEGER NOT NULL,
		NodeID     BLOB    NOT NULL,
		PeerID     TEXT    NOT NULL,
		LastSeen   INTEGER NOT NULL,
		PRIMARY KEY (bucket_idx, NodeID)
	);
	CREATE INDEX IF NOT EXISTS node_idx
		ON RoutingTable(bucket_idx, LastSeen);`

	// Create msgcert table + indexes
	if _, err := DB.Exec(createMsgCertTable); err != nil {
		return fmt.Errorf("creating msgcert table: %w", err)
	}

	// Create routing table + indexes
	if _, err := DB.Exec(createRoutingTable); err != nil {
		return fmt.Errorf("creating RoutingTable: %w", err)
	}

	return nil
}
