package repository

import (
	"context"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	MarkProcessed(ctx context.Context, id int64, status domain.PaymentStatus, failureReason string) (*domain.Payment, error)
}
