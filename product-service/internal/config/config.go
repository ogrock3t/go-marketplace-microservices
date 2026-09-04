package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func Load() *Config {
	return &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8081"),
		DatabaseDSN: getEnv("DATABASE_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
