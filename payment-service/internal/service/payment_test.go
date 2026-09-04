package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/dto"
	"payment-service/internal/events"
)

type paymentRepoMock struct {
	createErr error
	existing  *domain.Payment
	created   *domain.Payment
	processed *domain.Payment
}

func (m *paymentRepoMock) CreatePayment(_ context.Context, payment *domain.Payment) error {
	if m.createErr != nil {
		return m.createErr
	}

	payment.ID = 10
	payment.CreatedAt = time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	payment.UpdatedAt = payment.CreatedAt
	m.created = payment
	return nil
}

func (m *paymentRepoMock) GetPaymentByOrderID(_ context.Context, orderID int64) (*domain.Payment, error) {
	if m.existing == nil {
		return nil, domain.ErrPaymentNotFound
	}
	return m.existing, nil
}

func (m *paymentRepoMock) MarkProcessed(_ context.Context, id int64, status domain.PaymentStatus, failureReason string) (*domain.Payment, error) {
	processedAt := time.Date(2026, 9, 4, 12, 1, 0, 0, time.UTC)
	payment := *m.created
	if m.existing != nil {
		payment = *m.existing
	}
	payment.ID = id
	payment.Status = status
	payment.FailureReason = failureReason
	payment.ProcessedAt = &processedAt
	m.processed = &payment
	return &payment, nil
}

type publisherMock struct {
	eventType string
	payload   any
	err       error
}

func (m *publisherMock) Publish(_ context.Context, eventType string, payload any) error {
	m.eventType = eventType
	m.payload = payload
	return m.err
}

func TestProcessOrderCreatedCreatesPaymentAndPublishesSuccess(t *testing.T) {
	repo := &paymentRepoMock{}
	publisher := &publisherMock{}
	service := NewPaymentService(repo, publisher)

	payload, err := service.ProcessOrderCreated(context.Background(), dto.OrderCreatedPayload{
		ID:          100,
		UserID:      7,
		TotalAmount: 4500,
	})
	if err != nil {
		t.Fatalf("ProcessOrderCreated returned error: %v", err)
	}

	if repo.created == nil {
		t.Fatal("expected payment to be created")
	}
	if payload.Status != string(domain.StatusSuccess) || !payload.Success {
		t.Fatalf("expected successful payload, got %+v", payload)
	}
	if payload.OrderID != 100 || payload.UserID != 7 || payload.Amount != 4500 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if publisher.eventType != events.PaymentProcessedEvent {
		t.Fatalf("expected %s, got %s", events.PaymentProcessedEvent, publisher.eventType)
	}
}

func TestProcessOrderCreatedRepublishesExistingProcessedPayment(t *testing.T) {
	processedAt := time.Date(2026, 9, 4, 12, 1, 0, 0, time.UTC)
	repo := &paymentRepoMock{
		createErr: domain.ErrPaymentAlreadyExists,
		existing: &domain.Payment{
			ID:          10,
			OrderID:     100,
			UserID:      7,
			Amount:      4500,
			Status:      domain.StatusSuccess,
			ProcessedAt: &processedAt,
		},
	}
	publisher := &publisherMock{}
	service := NewPaymentService(repo, publisher)

	payload, err := service.ProcessOrderCreated(context.Background(), dto.OrderCreatedPayload{
		ID:          100,
		UserID:      7,
		TotalAmount: 4500,
	})
	if err != nil {
		t.Fatalf("ProcessOrderCreated returned error: %v", err)
	}

	if repo.processed != nil {
		t.Fatal("did not expect existing successful payment to be processed again")
	}
	if payload.PaymentID != 10 || !payload.Success {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if publisher.eventType != events.PaymentProcessedEvent {
		t.Fatalf("expected republished %s, got %s", events.PaymentProcessedEvent, publisher.eventType)
	}
}

func TestProcessOrderCreatedRejectsInvalidOrder(t *testing.T) {
	service := NewPaymentService(&paymentRepoMock{}, &publisherMock{})

	_, err := service.ProcessOrderCreated(context.Background(), dto.OrderCreatedPayload{
		ID:          100,
		UserID:      7,
		TotalAmount: 0,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProcessOrderCreatedReturnsPublishError(t *testing.T) {
	publishErr := errors.New("kafka failed")
	service := NewPaymentService(&paymentRepoMock{}, &publisherMock{err: publishErr})

	_, err := service.ProcessOrderCreated(context.Background(), dto.OrderCreatedPayload{
		ID:          100,
		UserID:      7,
		TotalAmount: 4500,
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("expected publish error, got %v", err)
	}
}
