package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusPaid      OrderStatus = "PAID"
	StatusCanceled  OrderStatus = "CANCELED"
	StatusCompleted OrderStatus = "COMPLETED"
)

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrInvalidOrderStatus       = errors.New("invalid order status")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrProductReservationFailed = errors.New("product reservation failed")
)

type Order struct {
	ID          int64       `db:"id"`
	UserID      int64       `db:"user_id"`
	Status      OrderStatus `db:"status"`
	TotalAmount int64       `db:"total_amount"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
	Items       []OrderItem
}

type OrderItem struct {
	ID        int64 `db:"id"`
	OrderID   int64 `db:"order_id"`
	ProductID int64 `db:"product_id"`
	Quantity  int64 `db:"quantity"`
	Price     int64 `db:"price"`
}

type ReservedProduct struct {
	ID    int64
	Price int64
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	switch OrderStatus(status) {
	case StatusCreated, StatusPaid, StatusCanceled, StatusCompleted:
		return OrderStatus(status), nil
	default:
		return "", ErrInvalidOrderStatus
	}
}

func CanTransition(from OrderStatus, to OrderStatus) bool {
	switch from {
	case StatusCreated:
		return to == StatusPaid || to == StatusCanceled
	case StatusPaid:
		return to == StatusCompleted || to == StatusCanceled
	case StatusCanceled, StatusCompleted:
		return false
	default:
		return false
	}
}
