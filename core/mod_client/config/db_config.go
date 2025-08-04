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

var DB *sql.DB

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
		return "./core/mod_client/data/libr.db"
	}
	return filepath.Join(filepath.Dir(ex), "core", "db", "data", "libr.db")
}

func createTables() error {
	createMsgModTable := `
	CREATE TABLE IF NOT EXISTS msgresult (
		sign TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		moderated INTEGER DEFAULT 0,
		modsign TEXT
	);
	CREATE INDEX IF NOT EXISTS indx_sign ON msgresult(sign);`

	_, err := DB.Exec(createMsgModTable)
	if err != nil {
		return fmt.Errorf("creating msgmod table: %w", err)
	}

	return nil
}
