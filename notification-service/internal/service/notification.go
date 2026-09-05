package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"notification-service/internal/dto"
)

type NotificationService struct {
	log *slog.Logger
}

func NewNotificationService(log *slog.Logger) *NotificationService {
	return &NotificationService{log: log}
}

func (s *NotificationService) NotifyOrderCreated(_ context.Context, order dto.OrderCreatedPayload) {
	s.log.Info("order notification sent",
		slog.String("message", orderCreatedMessage(order)),
		slog.Int64("order_id", order.ID),
		slog.Int64("user_id", order.UserID),
		slog.Int64("total_amount", order.TotalAmount),
	)
}

func (s *NotificationService) NotifyOrderStatusChanged(_ context.Context, event dto.OrderStatusChangedPayload) {
	message := orderStatusChangedMessage(event)
	s.log.Info("order status notification sent",
		slog.String("message", message),
		slog.Int64("order_id", event.OrderID),
		slog.Int64("user_id", event.UserID),
		slog.String("old_status", event.OldStatus),
		slog.String("new_status", event.NewStatus),
	)
}

func (s *NotificationService) NotifyPaymentProcessed(_ context.Context, payment dto.PaymentProcessedPayload) {
	message := paymentProcessedMessage(payment)
	s.log.Info("payment notification sent",
		slog.String("message", message),
		slog.Int64("payment_id", payment.PaymentID),
		slog.Int64("order_id", payment.OrderID),
		slog.Int64("user_id", payment.UserID),
		slog.Int64("amount", payment.Amount),
		slog.String("status", payment.Status),
		slog.Bool("success", payment.Success),
	)
}

func orderCreatedMessage(order dto.OrderCreatedPayload) string {
	return "Уведомление о создании заказа №" + int64ToString(order.ID) + " отправлено пользователю " + int64ToString(order.UserID)
}

func orderStatusChangedMessage(event dto.OrderStatusChangedPayload) string {
	switch strings.ToUpper(event.NewStatus) {
	case "CANCELED":
		return "Уведомление об отмене заказа №" + int64ToString(event.OrderID) + " отправлено пользователю " + int64ToString(event.UserID)
	case "COMPLETED":
		return "Уведомление о завершении заказа №" + int64ToString(event.OrderID) + " отправлено пользователю " + int64ToString(event.UserID)
	case "PAID":
		return "Уведомление об успешной оплате заказа №" + int64ToString(event.OrderID) + " отправлено пользователю " + int64ToString(event.UserID)
	default:
		return "Уведомление об изменении статуса заказа №" + int64ToString(event.OrderID) + " отправлено пользователю " + int64ToString(event.UserID)
	}
}

func paymentProcessedMessage(payment dto.PaymentProcessedPayload) string {
	if payment.Success || strings.EqualFold(payment.Status, "SUCCESS") {
		return "Уведомление об успешной оплате заказа №" + int64ToString(payment.OrderID) + " отправлено пользователю " + int64ToString(payment.UserID)
	}
	if payment.FailureReason != "" {
		return "Уведомление об ошибке оплаты заказа №" + int64ToString(payment.OrderID) + " отправлено пользователю " + int64ToString(payment.UserID) + ": " + payment.FailureReason
	}
	return "Уведомление об ошибке оплаты заказа №" + int64ToString(payment.OrderID) + " отправлено пользователю " + int64ToString(payment.UserID)
}

func int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}
