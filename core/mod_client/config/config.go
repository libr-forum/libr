package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

const K = 4
const Alpha = 4
const DeleteThreshold = 40.0
const MongoURI = "mongodb+srv://peer:peerhehe@cluster0.vswojqe.mongodb.net/" // Default MongoDB URI, can be overridden by environment variable

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
