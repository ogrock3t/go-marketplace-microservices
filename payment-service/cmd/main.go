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

	"payment-service/internal/config"
	"payment-service/internal/events"
	"payment-service/internal/service"
	storage "payment-service/internal/storage/postgres"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("Starting Payment Service")

	cfg := config.Load()
	if cfg.DatabaseDSN == "" {
		log.Error("DATABASE_DSN is required")
		os.Exit(1)
	}
	if len(cfg.KafkaBrokers) == 0 {
		log.Error("KAFKA_BROKERS is required")
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

	paymentStorage := storage.NewPaymentStorage(db)
	publisher := events.NewKafkaPublisher(cfg.KafkaBrokers, cfg.PaymentEventsTopic)
	defer publisher.Close()

	paymentService := service.NewPaymentService(paymentStorage, publisher)

	go events.RunOrderCreatedConsumer(
		ctx,
		cfg.KafkaBrokers,
		cfg.OrderEventsTopic,
		cfg.PaymentConsumerGroup,
		paymentService,
		log,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Info("starting health server", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
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
