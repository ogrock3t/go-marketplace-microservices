package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"order-service/internal/domain"
)

type OrderStorage struct {
	connect *Connection
}

func NewOrderStorage(connect *Connection) *OrderStorage {
	return &OrderStorage{connect: connect}
}

func (s *OrderStorage) CreateOrder(ctx context.Context, order *domain.Order) error {
	tx, err := s.connect.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin order transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const createOrderQuery = `
		INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, createOrderQuery, order.UserID, order.Status, order.TotalAmount).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order on postgres: %w", err)
	}

	const createItemQuery = `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		err = tx.QueryRow(
			ctx,
			createItemQuery,
			order.Items[i].OrderID,
			order.Items[i].ProductID,
			order.Items[i].Quantity,
			order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return fmt.Errorf("failed to create order item on postgres: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit order transaction: %w", err)
	}

	return nil
}

func (s *OrderStorage) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	const query = `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order
	err := s.connect.pool.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by id on postgres: %w", err)
	}

	items, err := s.listItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (s *OrderStorage) ListOrdersByUser(ctx context.Context, userID int64) ([]domain.Order, error) {
	const query = `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := s.connect.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by user on postgres: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}

		order.Items, err = s.listItemsByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate order rows: %w", err)
	}

	return orders, nil
}

func (s *OrderStorage) UpdateOrderStatus(ctx context.Context, id int64, status domain.OrderStatus) (*domain.Order, error) {
	const query = `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, status, total_amount, created_at, updated_at
	`

	var order domain.Order
	err := s.connect.pool.QueryRow(ctx, query, id, status).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to update order status on postgres: %w", err)
	}

	order.Items, err = s.listItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderStorage) listItemsByOrderID(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	const query = `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := s.connect.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list order items on postgres: %w", err)
	}
	defer rows.Close()

	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate order item rows: %w", err)
	}

	return items, nil
}
