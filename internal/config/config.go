package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RiotAPIKey string
	Region     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		RiotAPIKey: getEnv("RIOT_API_KEY", ""),
		Region:     getEnv("REGION", "asia"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
