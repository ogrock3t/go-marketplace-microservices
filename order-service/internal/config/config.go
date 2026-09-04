package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr             string
	DatabaseDSN          string
	ProductServiceURL    string
	RequestTimeout       time.Duration
	KafkaBrokers         string
	OrderEventsTopic     string
	PaymentEventsTopic   string
	PaymentConsumerGroup string
}

func Load() *Config {
	requestTimeout, err := time.ParseDuration(getEnv("REQUEST_TIMEOUT", "5s"))
	if err != nil {
		requestTimeout = 5 * time.Second
	}

	return &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8082"),
		DatabaseDSN:          getEnv("DATABASE_DSN", ""),
		ProductServiceURL:    getEnv("PRODUCT_SERVICE_URL", "http://product-service:8081"),
		RequestTimeout:       requestTimeout,
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
