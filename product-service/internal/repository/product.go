package repository

import (
	"context"
	"product-service/internal/domain"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	UpdateProductByID(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id int64) error
	ListProductsBySeller(ctx context.Context, sellerID int64) ([]*domain.Product, error)
	ListProductsByCategory(ctx context.Context, categoryID int64) ([]*domain.Product, error)
	ReserveProduct(ctx context.Context, id int64, quantity int64) (*domain.Product, error)
}
