package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service/internal/dto"
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

type OrderStatusUpdater interface {
	UpdateOrderStatus(ctx context.Context, id int64, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error)
}

type PaymentProcessedPayload struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Success bool   `json:"success"`
}

func RunPaymentProcessedConsumer(
	ctx context.Context,
	brokers []string,
	topic string,
	groupID string,
	orderService OrderStatusUpdater,
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
			log.Error("failed to read payment event", slog.String("error", err.Error()))
			continue
		}

		payload, err := decodePaymentProcessedPayload(msg.Value)
		if err != nil {
			log.Error("failed to decode payment event", slog.String("error", err.Error()))
			continue
		}
		if payload.OrderID == 0 || !isPaymentSuccess(payload) {
			continue
		}

		_, err = orderService.UpdateOrderStatus(ctx, payload.OrderID, &dto.UpdateOrderStatusRequest{
			Status: "PAID",
		})
		if err != nil {
			log.Error("failed to mark order as paid",
				slog.Int64("order_id", payload.OrderID),
				slog.String("error", err.Error()),
			)
			continue
		}

		log.Info("order marked as paid from payment event", slog.Int64("order_id", payload.OrderID))
	}
}

func decodePaymentProcessedPayload(raw []byte) (*PaymentProcessedPayload, error) {
	var envelope struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Payload) > 0 {
		raw = envelope.Payload
	}

	var payload PaymentProcessedPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

func isPaymentSuccess(payload *PaymentProcessedPayload) bool {
	status := strings.ToUpper(payload.Status)
	return payload.Success || status == "PAID" || status == "SUCCESS" || status == "PAYMENT_SUCCESS"
}
