package repository

import (
	"context"

	"order-service/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status domain.OrderStatus) (*domain.Order, error)
}
