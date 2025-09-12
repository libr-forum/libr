package config

import (
	"log"
	"os"
	"strconv"
	"embed"
	"github.com/joho/godotenv"
)
//go:embed .env
var _ embed.FS // Embedded .env file (unused, but required for go:embed syntax)



// Config holds all configuration for the application.
type Config struct {
	DBPath       string
	IP           string
	Port         int
	Bootstrap    string
	JSServerURL  string
	JSAPIKey     string
}

// Cfg is a global, public variable that holds the loaded configuration.



// Cfg is a global, public variable that holds the loaded configuration.
var Cfg Config
func init() {
	// Load the .env file.
	// It's okay if it fails, we can rely on OS environment variables as a fallback.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS environment variables")
	}

	// Load configuration into the Cfg struct
	Cfg = Config{
		DBPath:       getEnv("DB_PATH", "data/default.db"), // Provide a default
		IP:           getEnv("IP", "0.0.0.0"),             // Default to listen on all interfaces
		Port:         getEnvAsInt("PORT", 33122),         // Default port
		Bootstrap:    getEnv("BOOTSTRAP", ""),              // No default needed
		JSServerURL:  getEnv("JS_ServerURL", ""),           // No default needed
		JSAPIKey:     getEnv("JS_API_KEY", ""),             // No default needed
	}
	log.Println("Configuration loaded successfully")
}

// getEnv is a helper function to read an environment variable or return a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt is a helper to parse an environment variable as an integer.
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}