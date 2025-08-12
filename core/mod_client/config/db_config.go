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

	dbPath := filepath.Join(baseDir, "libr", "moddb", "mod.db")
	return dbPath
}

func createTables() error {
	createMsgModTable := `
	CREATE TABLE IF NOT EXISTS msgresult (
		sign TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		reason TEXT,
		moderated INTEGER,
		modsign TEXT
	);
	CREATE INDEX IF NOT EXISTS indx_sign ON msgresult(sign);`

	_, err := DB.Exec(createMsgModTable)
	if err != nil {
		return fmt.Errorf("creating msgmod table: %w", err)
	}

	return nil
}
