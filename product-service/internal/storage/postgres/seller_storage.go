package storage

import (
	"context"
	"fmt"
	"product-service/internal/domain"
)

type SellerStorage struct {
	connect *Connection
}

func NewSellerStorage(connect *Connection) *SellerStorage {
	return &SellerStorage{connect: connect}
}

func (s *SellerStorage) CreateSeller(ctx context.Context, seller *domain.Seller) error {
	const query = `
		INSERT INTO sellers (first_name, last_name, email)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.connect.pool.QueryRow(
		ctx,
		query,
		seller.FirstName,
		seller.LastName,
		seller.Email).Scan(&seller.ID, &seller.CreatedAt, &seller.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create seller on postgresql: %w", err)
	}

	return nil
}

func (s *SellerStorage) GetSellerByID(ctx context.Context, id int64) (*domain.Seller, error) {
	const query = `
		SELECT id, first_name, last_name, email, created_at, updated_at
		FROM sellers
		WHERE id = $1
	`

	var seller domain.Seller

	err := s.connect.pool.QueryRow(ctx, query, id).Scan(
		&seller.ID,
		&seller.FirstName,
		&seller.LastName,
		&seller.Email,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrSellerNotFound
	}

	return &seller, nil
}

func (s *SellerStorage) UpdateSeller(ctx context.Context, seller *domain.Seller) error {
	const query = `
		UPDATE sellers
		SET first_name = $1, last_name = $2, email = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := s.connect.pool.Exec(
		ctx,
		query,
		seller.FirstName,
		seller.LastName,
		seller.Email,
		seller.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update seller on postgresql: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrSellerNotFound
	}

	return nil
}

func (s *SellerStorage) DeleteSeller(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM sellers
		WHERE id = $1
	`

	result, err := s.connect.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete seller on postgresql: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrSellerNotFound
	}

	return nil
}
