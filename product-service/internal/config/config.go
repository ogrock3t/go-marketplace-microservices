package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr        string
	GRPCAddr        string
	DatabaseDSN     string
	ShutdownTimeout time.Duration
}

func Load() *Config {
	shutdownTimeout, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		shutdownTimeout = 10 * time.Second
	}

	return &Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8081"),
		GRPCAddr:        getEnv("GRPC_ADDR", ":9091"),
		DatabaseDSN:     getEnv("DATABASE_DSN", ""),
		ShutdownTimeout: shutdownTimeout,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
