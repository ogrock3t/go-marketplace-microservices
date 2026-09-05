package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/config"
	"notification-service/internal/events"
	"notification-service/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("Starting Notification Service")

	cfg := config.Load()
	if len(cfg.KafkaBrokers) == 0 {
		log.Error("KAFKA_BROKERS is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	notificationService := service.NewNotificationService(log)

	consumer := events.NewConsumer(
		cfg.KafkaBrokers,
		[]string{cfg.OrderEventsTopic, cfg.PaymentEventsTopic},
		cfg.ConsumerGroup,
		notificationService,
		log,
	)

	go consumer.Run(ctx)

	log.Info("notification consumers started",
		slog.String("orders_topic", cfg.OrderEventsTopic),
		slog.String("payments_topic", cfg.PaymentEventsTopic),
		slog.String("group", cfg.ConsumerGroup),
	)

	<-ctx.Done()
	log.Info("Notification Service stopped")
}
