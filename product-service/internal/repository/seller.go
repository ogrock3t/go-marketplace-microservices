package repository

import (
	"context"
	"product-service/internal/domain"
)

type SellerRepository interface {
	CreateSeller(ctx context.Context, seller *domain.Seller) error
	GetSellerByID(ctx context.Context, id int64) (*domain.Seller, error)
	UpdateSeller(ctx context.Context, seller *domain.Seller) error
	DeleteSeller(ctx context.Context, id int64) error
}
