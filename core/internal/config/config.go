package config

import (
	"errors"
	"os"
	"strconv"

	"stock-market/backend/internal/auth"

	"github.com/joho/godotenv"
)

const (
	defaultPort                         = "8080"
	defaultRefreshTokenCookie           = "refreshToken"
	defaultAccessTokenExpiresInSeconds  = 60
	defaultRefreshTokenExpiresInSeconds = 604800
)

type Config struct {
	Port          string
	FinnhubAPIKey string
	MongoURI      string
	RedisURL      string
	Telegram      TelegramConfig
	Auth          auth.Config
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

func (c TelegramConfig) Enabled() bool {
	return c.BotToken != "" && c.ChatID != ""
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		return nil, errors.New("FINNHUB_API_KEY environment variable is required")
	}

	mongoURI := os.Getenv("MONGO_CONNECTION_STRING")
	if mongoURI == "" {
		return nil, errors.New("MONGO_CONNECTION_STRING environment variable is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, errors.New("REDIS_URL environment variable is required")
	}

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return nil, errors.New("JWT_ACCESS_SECRET environment variable is required")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, errors.New("JWT_REFRESH_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	refreshCookie := os.Getenv("REFRESH_TOKEN_COOKIE")
	if refreshCookie == "" {
		refreshCookie = defaultRefreshTokenCookie
	}

	accessTTL, err := parsePositiveIntEnv("JWT_ACCESS_EXPIRES_IN_SECONDS", defaultAccessTokenExpiresInSeconds)
	if err != nil {
		return nil, err
	}

	refreshTTL, err := parsePositiveIntEnv("JWT_REFRESH_EXPIRES_IN_SECONDS", defaultRefreshTokenExpiresInSeconds)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:          port,
		FinnhubAPIKey: apiKey,
		MongoURI:      mongoURI,
		RedisURL:      redisURL,
		Telegram: TelegramConfig{
			BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		},
		Auth: auth.Config{
			RefreshTokenCookie:           refreshCookie,
			AccessJWTSecret:              accessSecret,
			RefreshJWTSecret:             refreshSecret,
			AccessTokenExpiresInSeconds:  accessTTL,
			RefreshTokenExpiresInSeconds: refreshTTL,
			SecureCookies:                os.Getenv("NODE_ENV") == "production",
		},
	}, nil
}

func parsePositiveIntEnv(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New(key + " must be a positive integer")
	}

	return value, nil
}
