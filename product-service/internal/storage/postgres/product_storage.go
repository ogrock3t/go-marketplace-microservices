package storage

import (
	"context"
	"errors"
	"fmt"
	"product-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

var (
	ErrScanProduct        = fmt.Errorf("failed to scan product row")
	ErrIterateProductRows = fmt.Errorf("error iterating over product rows")
)

type ProductStorage struct {
	connect *Connection
}

func NewProductStorage(connect *Connection) *ProductStorage {
	return &ProductStorage{connect: connect}
}

func (s *ProductStorage) CreateProduct(ctx context.Context, product *domain.Product) error {
	const query = `
		INSERT INTO products (seller_id, category_id, name, description, price, available_quantity, status)
		values ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := s.connect.pool.QueryRow(
		ctx,
		query,
		product.SellerID,
		product.CategoryID,
		product.Name,
		product.Description,
		product.Price,
		product.AvailableQuantity,
		product.Status).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create a product on postgres: %w", err)
	}

	return nil
}

func (s *ProductStorage) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	const query = `
		SELECT id, seller_id, category_id, name, description, price, available_quantity, status, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product domain.Product

	err := s.connect.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.AvailableQuantity,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by id on postgres: %w", err)
	}

	return &product, nil
}

func (s *ProductStorage) UpdateProductByID(ctx context.Context, product *domain.Product) error {
	const query = `
		UPDATE products
		SET seller_id = $1, category_id = $2, name = $3, description = $4, price = $5, available_quantity = $6, status = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := s.connect.pool.Exec(
		ctx,
		query,
		product.SellerID,
		product.CategoryID,
		product.Name,
		product.Description,
		product.Price,
		product.AvailableQuantity,
		product.Status,
		product.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update a product on postgres: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (s *ProductStorage) DeleteProduct(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM products
		WHERE id = $1
	`

	result, err := s.connect.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete a product on postgres: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (s *ProductStorage) ListProductsBySeller(ctx context.Context, sellerID int64) ([]*domain.Product, error) {
	const query = `
		SELECT id, seller_id, category_id, name, description, price, available_quantity, status, created_at, updated_at
		FROM products
		WHERE seller_id = $1
		ORDER BY id
	`

	rows, err := s.connect.pool.Query(ctx, query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by seller on postgres: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)

	for rows.Next() {
		var product domain.Product

		err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.CategoryID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.AvailableQuantity,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrIterateProductRows
	}

	return products, nil
}

func (s *ProductStorage) ListProductsByCategory(ctx context.Context, categoryID int64) ([]*domain.Product, error) {
	const query = `
		SELECT id, seller_id, category_id, name, description, price, available_quantity, status, created_at, updated_at
		FROM products
		WHERE category_id = $1
		ORDER BY id
	`

	rows, err := s.connect.pool.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by category on postgres: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)

	for rows.Next() {
		var product domain.Product

		err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.CategoryID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.AvailableQuantity,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrIterateProductRows
	}

	return products, nil
}

func (s *ProductStorage) ReserveProduct(ctx context.Context, id int64, quantity int64) (*domain.Product, error) {
	const query = `
		UPDATE products
		SET
			available_quantity = available_quantity - $2,
			status = CASE
				WHEN available_quantity - $2 = 0 THEN $3
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = $1
			AND status = $4
			AND available_quantity >= $2
		RETURNING id, seller_id, category_id, name, description, price, available_quantity, status, created_at, updated_at
	`

	var product domain.Product
	err := s.connect.pool.QueryRow(ctx, query, id, quantity, domain.StatusOutOfStock, domain.StatusActive).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.AvailableQuantity,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == nil {
		return &product, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to reserve product on postgres: %w", err)
	}

	if _, getErr := s.GetProductByID(ctx, id); getErr != nil {
		return nil, getErr
	}

	return nil, domain.ErrInsufficientStock
}

func (s *ProductStorage) ReleaseProduct(ctx context.Context, id int64, quantity int64) (*domain.Product, error) {
	const query = `
		UPDATE products
		SET
			available_quantity = available_quantity + $2,
			status = CASE
				WHEN status = $3 THEN $4
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, seller_id, category_id, name, description, price, available_quantity, status, created_at, updated_at
	`

	var product domain.Product
	err := s.connect.pool.QueryRow(ctx, query, id, quantity, domain.StatusOutOfStock, domain.StatusActive).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.AvailableQuantity,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to release product on postgres: %w", err)
	}

	return &product, nil
}
