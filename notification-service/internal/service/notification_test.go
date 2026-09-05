package service

import (
	"testing"

	"notification-service/internal/dto"
)

func TestPaymentProcessedMessageSuccess(t *testing.T) {
	message := paymentProcessedMessage(dto.PaymentProcessedPayload{
		OrderID: 100,
		UserID:  7,
		Status:  "SUCCESS",
		Success: true,
	})

	expected := "Уведомление об успешной оплате заказа №100 отправлено пользователю 7"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

func TestPaymentProcessedMessageFailure(t *testing.T) {
	message := paymentProcessedMessage(dto.PaymentProcessedPayload{
		OrderID:       100,
		UserID:        7,
		Status:        "FAILED",
		FailureReason: "card declined",
	})

	expected := "Уведомление об ошибке оплаты заказа №100 отправлено пользователю 7: card declined"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

func TestOrderStatusChangedMessageCanceled(t *testing.T) {
	message := orderStatusChangedMessage(dto.OrderStatusChangedPayload{
		OrderID:   100,
		UserID:    7,
		OldStatus: "CREATED",
		NewStatus: "CANCELED",
	})

	expected := "Уведомление об отмене заказа №100 отправлено пользователю 7"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}
