package repository

import (
	"context"
	"errors"
	"product-service/internal/domain"
)

var (
	ErrorCategoryNotFound = errors.New("category not found")
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
	GetCategoryByID(ctx context.Context, id int64) (*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int64) error

	// ListCategories returns only root categories (those parent_id is null)
	ListCategories(ctx context.Context) ([]*domain.Category, error)

	// ListSubcategories returns subcategories of a given parent category
	ListSubcategories(ctx context.Context, parentID int64) ([]*domain.Category, error)
}
