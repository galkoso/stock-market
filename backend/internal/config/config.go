package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultPort = "8080"
)

type Config struct {
	Port          string
	FinnhubAPIKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		return nil, errors.New("FINNHUB_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return &Config{
		Port:          port,
		FinnhubAPIKey: apiKey,
	}, nil
}
