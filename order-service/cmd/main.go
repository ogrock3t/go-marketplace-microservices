package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	productclient "order-service/internal/client/product"
	"order-service/internal/config"
	"order-service/internal/events"
	httpserver "order-service/internal/http-server"
	"order-service/internal/http-server/handler"
	"order-service/internal/service"
	storage "order-service/internal/storage/postgres"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("Starting Order Service")

	cfg := config.Load()
	if cfg.DatabaseDSN == "" {
		log.Error("DATABASE_DSN is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://migrations", migrationDSN(cfg.DatabaseDSN))
	if err != nil {
		log.Error("failed to init migrations", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.NewConnection(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	orderStorage := storage.NewOrderStorage(db)
	productClient := productclient.NewClient(cfg.ProductServiceURL, &http.Client{Timeout: cfg.RequestTimeout})
	publisher := events.Publisher(events.NewLogPublisher(log))
	if cfg.KafkaBrokers != "" {
		kafkaBrokers := splitCSV(cfg.KafkaBrokers)
		kafkaPublisher := events.NewKafkaPublisher(kafkaBrokers, cfg.OrderEventsTopic)
		defer kafkaPublisher.Close()
		publisher = kafkaPublisher
		log.Info("kafka publisher initialized", "topic", cfg.OrderEventsTopic)
	}

	orderService := service.NewOrderService(orderStorage, productClient, publisher)
	orderHandler := handler.NewOrderHandler(orderService)
	router := httpserver.NewRouter(orderHandler, log)

	if cfg.KafkaBrokers != "" {
		go events.RunPaymentProcessedConsumer(
			ctx,
			splitCSV(cfg.KafkaBrokers),
			cfg.PaymentEventsTopic,
			cfg.PaymentConsumerGroup,
			orderService,
			log,
		)
	}

	log.Info("starting server", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
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

func migrationDSN(dsn string) string {
	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	case strings.HasPrefix(dsn, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://")
	default:
		return dsn
	}
}
