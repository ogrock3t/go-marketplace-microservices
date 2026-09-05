package events

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"notification-service/internal/dto"
)

type notificationHandlerMock struct {
	orderCreated       *dto.OrderCreatedPayload
	orderStatusChanged *dto.OrderStatusChangedPayload
	paymentProcessed   *dto.PaymentProcessedPayload
}

func (m *notificationHandlerMock) NotifyOrderCreated(_ context.Context, order dto.OrderCreatedPayload) {
	m.orderCreated = &order
}

func (m *notificationHandlerMock) NotifyOrderStatusChanged(_ context.Context, event dto.OrderStatusChangedPayload) {
	m.orderStatusChanged = &event
}

func (m *notificationHandlerMock) NotifyPaymentProcessed(_ context.Context, payment dto.PaymentProcessedPayload) {
	m.paymentProcessed = &payment
}

func TestHandleMessageRoutesPaymentProcessedEvent(t *testing.T) {
	handler := &notificationHandlerMock{}
	consumer := &Consumer{
		handler: handler,
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	raw := mustMarshal(t, rawEnvelope{
		Type: PaymentProcessedEvent,
		Payload: json.RawMessage(mustMarshal(t, dto.PaymentProcessedPayload{
			PaymentID: 10,
			OrderID:   100,
			UserID:    7,
			Amount:    4500,
			Status:    "SUCCESS",
			Success:   true,
		})),
	})

	if err := consumer.handleMessage(context.Background(), raw); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}
	if handler.paymentProcessed == nil {
		t.Fatal("expected payment notification")
	}
	if handler.paymentProcessed.OrderID != 100 || !handler.paymentProcessed.Success {
		t.Fatalf("unexpected payment payload: %+v", handler.paymentProcessed)
	}
}

func TestHandleMessageRoutesOrderStatusChangedEvent(t *testing.T) {
	handler := &notificationHandlerMock{}
	consumer := &Consumer{
		handler: handler,
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	raw := mustMarshal(t, rawEnvelope{
		Type: OrderStatusChangedEvent,
		Payload: json.RawMessage(mustMarshal(t, dto.OrderStatusChangedPayload{
			OrderID:   100,
			UserID:    7,
			OldStatus: "CREATED",
			NewStatus: "CANCELED",
		})),
	})

	if err := consumer.handleMessage(context.Background(), raw); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}
	if handler.orderStatusChanged == nil {
		t.Fatal("expected order status notification")
	}
	if handler.orderStatusChanged.NewStatus != "CANCELED" {
		t.Fatalf("unexpected status payload: %+v", handler.orderStatusChanged)
	}
}

func TestHandleMessageIgnoresUnknownEvent(t *testing.T) {
	handler := &notificationHandlerMock{}
	consumer := &Consumer{
		handler: handler,
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	raw := mustMarshal(t, rawEnvelope{
		Type:    "UnknownEvent",
		Payload: json.RawMessage(mustMarshal(t, map[string]string{"ignored": "true"})),
	})

	if err := consumer.handleMessage(context.Background(), raw); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}
	if handler.orderCreated != nil || handler.orderStatusChanged != nil || handler.paymentProcessed != nil {
		t.Fatalf("expected no notifications, got %+v", handler)
	}
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal value: %v", err)
	}
	return raw
}
