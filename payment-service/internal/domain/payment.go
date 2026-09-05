package domain

import (
	"errors"
	"time"
)

type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists")
)

type Payment struct {
	ID            int64         `db:"id"`
	OrderID       int64         `db:"order_id"`
	UserID        int64         `db:"user_id"`
	Amount        int64         `db:"amount"`
	Status        PaymentStatus `db:"status"`
	FailureReason string        `db:"failure_reason"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
	ProcessedAt   *time.Time    `db:"processed_at"`
}
