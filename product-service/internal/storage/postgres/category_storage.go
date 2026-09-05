package storage

import (
	"context"
	"fmt"
	"product-service/internal/domain"
)

var (
	ErrScanCategory        = fmt.Errorf("failed to scan category row")
	ErrIterateCategoryRows = fmt.Errorf("error iterating over category rows")
)

type CategoryStorage struct {
	connect *Connection
}

func NewCategoryStorage(connect *Connection) *CategoryStorage {
	return &CategoryStorage{connect: connect}
}

func (s *CategoryStorage) CreateCategory(ctx context.Context, category *domain.Category) error {
	const query = `
		INSERT INTO categories (parent_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.connect.pool.QueryRow(
		ctx,
		query,
		category.ParentID,
		category.Name,
		category.Description).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create a category on postgres: %w", err)
	}

	return nil
}

func (s *CategoryStorage) GetCategoryByID(ctx context.Context, id int64) (*domain.Category, error) {
	const query = `
		SELECT id, parent_id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category domain.Category
	err := s.connect.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.ParentID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrCategoryNotFound
	}

	return &category, nil
}

func (s *CategoryStorage) UpdateCategory(ctx context.Context, category *domain.Category) error {
	const query = `
		UPDATE categories
		SET parent_id = $1, name = $2, description = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	result, err := s.connect.pool.Exec(
		ctx,
		query,
		category.ParentID,
		category.Name,
		category.Description,
		category.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update category on postgres: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func (s *CategoryStorage) DeleteCategory(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM categories
		WHERE id = $1
	`

	result, err := s.connect.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category on postgres: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func (s *CategoryStorage) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	const query = `
		SELECT id, parent_id, name, description, created_at, updated_at
		FROM categories
		WHERE parent_id IS NULL
	`

	rows, err := s.connect.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories on postgres: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.Category, 0)

	for rows.Next() {
		var category domain.Category

		err := rows.Scan(
			&category.ID,
			&category.ParentID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, ErrScanCategory
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrIterateCategoryRows
	}

	return categories, nil
}

func (s *CategoryStorage) ListSubcategories(ctx context.Context, parentID int64) ([]*domain.Category, error) {
	const query = `
		SELECT id, parent_id, name, description, created_at, updated_at
		FROM categories
		WHERE parent_id = $1
	`

	rows, err := s.connect.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subcategories on postgres: %w", err)
	}
	defer rows.Close()

	subcategories := make([]*domain.Category, 0)

	for rows.Next() {
		var category domain.Category

		err := rows.Scan(
			&category.ID,
			&category.ParentID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, ErrScanCategory
		}

		subcategories = append(subcategories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrIterateCategoryRows
	}

	return subcategories, nil
}
