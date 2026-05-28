package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr            string
	DatabaseDSN         string
	RSAPrivateKeyPath   string
	AccessTokenDuration time.Duration
}

func Load() *Config {
	accessTokenDuration, err := time.ParseDuration(getEnv("ACCESS_TOKEN_DURATION", "15m"))
	if err != nil {
		accessTokenDuration = 15 * time.Minute
	}

	return &Config{
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		DatabaseDSN:         getEnv("DATABASE_DSN", ""),
		RSAPrivateKeyPath:   getEnv("RSA_PRIVATE_KEY_PATH", ""),
		AccessTokenDuration: accessTokenDuration,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
