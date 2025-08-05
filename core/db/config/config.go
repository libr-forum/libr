package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const K = 4
const RepMajority = 0.5

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

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite DB: %v", err)
	}

	// Optional: enable foreign keys and WAL mode
	_, _ = DB.Exec("PRAGMA foreign_keys = ON;")
	_, _ = DB.Exec("PRAGMA journal_mode = WAL;")

	err = createTables()
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	log.Println("SQLite database initialized successfully.")
}

func getDBPath() string {
	// Check if DB_PATH is set in the environment
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		fmt.Println(envPath)
		return envPath
	}

	// Otherwise, fallback to a portable path next to the executable
	ex, err := os.Executable()
	if err != nil {
		return "./core/db/data/libr.db" // fallback for dev/debug
	}
	return filepath.Join(filepath.Dir(ex), "core", "db", "data", "libr.db")
}

func createTables() error {
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rt TEXT NOT NULL
	);`

	_, err := DB.Exec(createMsgCertTable)
	if err != nil {
		return fmt.Errorf("creating msgcert table: %w", err)
	}

	_, err = DB.Exec(createRoutingTable)
	if err != nil {
		return fmt.Errorf("creating RoutingTable: %w", err)
	}

	return nil
}
