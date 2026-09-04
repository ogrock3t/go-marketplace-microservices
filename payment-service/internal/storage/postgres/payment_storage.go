package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"payment-service/internal/domain"
)

type PaymentStorage struct {
	connect *Connection
}

func NewPaymentStorage(connect *Connection) *PaymentStorage {
	return &PaymentStorage{connect: connect}
}

func (s *PaymentStorage) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	const query = `
		INSERT INTO payments (order_id, user_id, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := s.connect.pool.QueryRow(
		ctx,
		query,
		payment.OrderID,
		payment.UserID,
		payment.Amount,
		payment.Status,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPaymentAlreadyExists
		}
		return fmt.Errorf("failed to create payment on postgres: %w", err)
	}

	return nil
}

func (s *PaymentStorage) GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, user_id, amount, status, failure_reason, created_at, updated_at, processed_at
		FROM payments
		WHERE order_id = $1
	`

	var payment domain.Payment
	err := s.connect.pool.QueryRow(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Status,
		&payment.FailureReason,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by order id on postgres: %w", err)
	}

	return &payment, nil
}

func (s *PaymentStorage) MarkProcessed(ctx context.Context, id int64, status domain.PaymentStatus, failureReason string) (*domain.Payment, error) {
	const query = `
		UPDATE payments
		SET status = $2, failure_reason = $3, processed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, order_id, user_id, amount, status, failure_reason, created_at, updated_at, processed_at
	`

	var payment domain.Payment
	err := s.connect.pool.QueryRow(ctx, query, id, status, failureReason).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Status,
		&payment.FailureReason,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to mark payment processed on postgres: %w", err)
	}

	return &payment, nil
}
