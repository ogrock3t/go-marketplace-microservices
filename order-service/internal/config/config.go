package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr             string
	DatabaseDSN          string
	ProductGRPCAddr      string
	ShutdownTimeout      time.Duration
	KafkaBrokers         string
	OrderEventsTopic     string
	PaymentEventsTopic   string
	PaymentConsumerGroup string
}

func Load() *Config {
	shutdownTimeout, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		shutdownTimeout = 10 * time.Second
	}

	return &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8082"),
		DatabaseDSN:          getEnv("DATABASE_DSN", ""),
		ProductGRPCAddr:      getEnv("PRODUCT_GRPC_ADDR", "product-service:9091"),
		ShutdownTimeout:      shutdownTimeout,
		KafkaBrokers:         getEnv("KAFKA_BROKERS", ""),
		OrderEventsTopic:     getEnv("ORDER_EVENTS_TOPIC", "orders"),
		PaymentEventsTopic:   getEnv("PAYMENT_EVENTS_TOPIC", "payments"),
		PaymentConsumerGroup: getEnv("PAYMENT_CONSUMER_GROUP", "order-service"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
