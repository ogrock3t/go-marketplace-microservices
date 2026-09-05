package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string
	DatabaseDSN          string
	KafkaBrokers         []string
	OrderEventsTopic     string
	PaymentEventsTopic   string
	PaymentConsumerGroup string
	ShutdownTimeout      time.Duration
}

func Load() *Config {
	shutdownTimeout, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		shutdownTimeout = 10 * time.Second
	}

	return &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8083"),
		DatabaseDSN:          getEnv("DATABASE_DSN", ""),
		KafkaBrokers:         splitCSV(getEnv("KAFKA_BROKERS", "kafka:9092")),
		OrderEventsTopic:     getEnv("ORDER_EVENTS_TOPIC", "orders"),
		PaymentEventsTopic:   getEnv("PAYMENT_EVENTS_TOPIC", "payments"),
		PaymentConsumerGroup: getEnv("PAYMENT_CONSUMER_GROUP", "payment-service"),
		ShutdownTimeout:      shutdownTimeout,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
