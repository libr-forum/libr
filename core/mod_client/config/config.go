package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

const K = 4
const Alpha = 4
const DeleteThreshold = 40.0

// ServerURL is the base URL of the librserver discovery server.
// Override with the SERVER_URL environment variable.
const ServerURL = "https://libr-relay007-1.onrender.com"

// RegistryRefreshSeconds is how often (in seconds) a mod refreshes its
// presence in the discovery server. Should be less than the server-side TTL.
const RegistryRefreshSeconds = 90

// GetServerURL returns the discovery server URL, preferring the SERVER_URL
// environment variable over the compiled-in default.
func GetServerURL() string {
	if u := os.Getenv("SERVER_URL"); u != "" {
		return u
	}
	return ServerURL
}

type Config struct {

	// External API keys
	GEMINI_API_KEY string `env:"GEMINI_API_KEY"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file loaded (production mode?)")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
