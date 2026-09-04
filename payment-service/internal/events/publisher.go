package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

const (
	OrderCreatedEvent     = "OrderCreatedEvent"
	PaymentProcessedEvent = "PaymentProcessedEvent"
)

type Publisher interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

type Envelope struct {
	Type       string `json:"type"`
	Payload    any    `json:"payload"`
	OccurredAt string `json:"occurred_at"`
}

type LogPublisher struct {
	log *slog.Logger
}

func NewLogPublisher(log *slog.Logger) *LogPublisher {
	return &LogPublisher{log: log}
}

func (p *LogPublisher) Publish(_ context.Context, eventType string, payload any) error {
	raw, err := json.Marshal(Envelope{
		Type:       eventType,
		Payload:    payload,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	p.log.Info("event published", slog.String("type", eventType), slog.String("payload", string(raw)))
	return nil
}
