package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"notification-service/internal/dto"
)

const (
	OrderCreatedEvent       = "OrderCreatedEvent"
	OrderStatusChangedEvent = "OrderStatusChangedEvent"
	PaymentProcessedEvent   = "PaymentProcessedEvent"
)

type NotificationHandler interface {
	NotifyOrderCreated(ctx context.Context, order dto.OrderCreatedPayload)
	NotifyOrderStatusChanged(ctx context.Context, event dto.OrderStatusChangedPayload)
	NotifyPaymentProcessed(ctx context.Context, payment dto.PaymentProcessedPayload)
}

type Consumer struct {
	reader  *kafka.Reader
	handler NotificationHandler
	log     *slog.Logger
}

func NewConsumer(brokers []string, topics []string, groupID string, handler NotificationHandler, log *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupTopics: topics,
			GroupID:     groupID,
		}),
		handler: handler,
		log:     log,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	defer c.reader.Close()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			c.log.Error("failed to read kafka message", slog.String("error", err.Error()))
			continue
		}

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			c.log.Error("failed to handle notification event", slog.String("error", err.Error()))
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, raw []byte) error {
	envelope, err := decodeEnvelope(raw)
	if err != nil {
		return err
	}

	switch envelope.Type {
	case OrderCreatedEvent:
		var payload dto.OrderCreatedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("failed to decode order created payload: %w", err)
		}
		c.handler.NotifyOrderCreated(ctx, payload)
	case OrderStatusChangedEvent:
		var payload dto.OrderStatusChangedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("failed to decode order status changed payload: %w", err)
		}
		c.handler.NotifyOrderStatusChanged(ctx, payload)
	case PaymentProcessedEvent:
		var payload dto.PaymentProcessedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("failed to decode payment processed payload: %w", err)
		}
		c.handler.NotifyPaymentProcessed(ctx, payload)
	default:
		c.log.Debug("ignored kafka event", slog.String("type", envelope.Type))
	}

	return nil
}

type rawEnvelope struct {
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt string          `json:"occurred_at"`
}

func decodeEnvelope(raw []byte) (*rawEnvelope, error) {
	var envelope rawEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode envelope: %w", err)
	}
	if envelope.Type == "" {
		return nil, errors.New("event type is required")
	}
	if len(envelope.Payload) == 0 {
		return nil, errors.New("event payload is required")
	}

	return &envelope, nil
}
