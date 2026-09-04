package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"payment-service/internal/dto"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, eventType string, payload any) error {
	raw, err := json.Marshal(Envelope{
		Type:       eventType,
		Payload:    payload,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to encode kafka event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventType),
		Value: raw,
		Time:  time.Now().UTC(),
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

type OrderCreatedProcessor interface {
	ProcessOrderCreated(ctx context.Context, order dto.OrderCreatedPayload) (*dto.PaymentProcessedPayload, error)
}

func RunOrderCreatedConsumer(
	ctx context.Context,
	brokers []string,
	topic string,
	groupID string,
	processor OrderCreatedProcessor,
	log *slog.Logger,
) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Error("failed to read order event", slog.String("error", err.Error()))
			continue
		}

		order, ok, err := decodeOrderCreatedPayload(msg.Value)
		if err != nil {
			log.Error("failed to decode order event", slog.String("error", err.Error()))
			continue
		}
		if !ok {
			continue
		}

		payment, err := processor.ProcessOrderCreated(ctx, *order)
		if err != nil {
			log.Error("failed to process payment",
				slog.Int64("order_id", order.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		log.Info("payment processed",
			slog.Int64("payment_id", payment.PaymentID),
			slog.Int64("order_id", payment.OrderID),
			slog.String("status", payment.Status),
		)
	}
}

func decodeOrderCreatedPayload(raw []byte) (*dto.OrderCreatedPayload, bool, error) {
	var envelope struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Type != "" {
		if envelope.Type != OrderCreatedEvent {
			return nil, false, nil
		}
		raw = envelope.Payload
	}

	var order dto.OrderCreatedPayload
	if err := json.Unmarshal(raw, &order); err != nil {
		return nil, false, err
	}

	return &order, true, nil
}
