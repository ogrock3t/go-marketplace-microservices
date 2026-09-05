package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers       []string
	OrderEventsTopic   string
	PaymentEventsTopic string
	ConsumerGroup      string
}

func Load() *Config {
	return &Config{
		KafkaBrokers:       splitCSV(getEnv("KAFKA_BROKERS", "kafka:9092")),
		OrderEventsTopic:   getEnv("ORDER_EVENTS_TOPIC", "orders"),
		PaymentEventsTopic: getEnv("PAYMENT_EVENTS_TOPIC", "payments"),
		ConsumerGroup:      getEnv("NOTIFICATION_CONSUMER_GROUP", "notification-service"),
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
