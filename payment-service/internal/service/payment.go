package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/dto"
	"payment-service/internal/events"
	"payment-service/internal/repository"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepository
	publisher   events.Publisher
}

func NewPaymentService(paymentRepo repository.PaymentRepository, publisher events.Publisher) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		publisher:   publisher,
	}
}

func (s *PaymentService) ProcessOrderCreated(ctx context.Context, order dto.OrderCreatedPayload) (*dto.PaymentProcessedPayload, error) {
	if order.ID <= 0 {
		return nil, errors.New("order id is required")
	}
	if order.UserID <= 0 {
		return nil, errors.New("user id is required")
	}
	if order.TotalAmount <= 0 {
		return nil, errors.New("total amount must be positive")
	}

	payment := &domain.Payment{
		OrderID: order.ID,
		UserID:  order.UserID,
		Amount:  order.TotalAmount,
		Status:  domain.StatusPending,
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		if !errors.Is(err, domain.ErrPaymentAlreadyExists) {
			return nil, err
		}

		existing, getErr := s.paymentRepo.GetPaymentByOrderID(ctx, order.ID)
		if getErr != nil {
			return nil, getErr
		}
		if existing.Status != domain.StatusPending {
			payload := paymentToProcessedPayload(existing)
			if err := s.publisher.Publish(ctx, events.PaymentProcessedEvent, payload); err != nil {
				return nil, fmt.Errorf("failed to publish existing payment processed event: %w", err)
			}
			return payload, nil
		}
		payment = existing
	}

	processed, err := s.paymentRepo.MarkProcessed(ctx, payment.ID, domain.StatusSuccess, "")
	if err != nil {
		return nil, err
	}

	payload := paymentToProcessedPayload(processed)
	if err := s.publisher.Publish(ctx, events.PaymentProcessedEvent, payload); err != nil {
		return nil, fmt.Errorf("failed to publish payment processed event: %w", err)
	}

	return payload, nil
}

func paymentToProcessedPayload(payment *domain.Payment) *dto.PaymentProcessedPayload {
	processedAt := time.Now().UTC().Format(time.RFC3339)
	if payment.ProcessedAt != nil {
		processedAt = payment.ProcessedAt.UTC().Format(time.RFC3339)
	}

	return &dto.PaymentProcessedPayload{
		PaymentID:     payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Status:        string(payment.Status),
		Success:       payment.Status == domain.StatusSuccess,
		FailureReason: payment.FailureReason,
		ProcessedAt:   processedAt,
	}
}
