package service

import (
	"context"
	"testing"

	"product-service/internal/domain"
	"product-service/internal/dto"
)

type productRepoMock struct {
	product          *domain.Product
	reservedID       int64
	reservedQuantity int64
}

func (m *productRepoMock) CreateProduct(_ context.Context, product *domain.Product) error {
	product.ID = 10
	m.product = product
	return nil
}

func (m *productRepoMock) GetProductByID(_ context.Context, id int64) (*domain.Product, error) {
	return m.product, nil
}

func (m *productRepoMock) UpdateProductByID(_ context.Context, product *domain.Product) error {
	m.product = product
	return nil
}

func (m *productRepoMock) DeleteProduct(_ context.Context, id int64) error {
	return nil
}

func (m *productRepoMock) ListProductsBySeller(_ context.Context, sellerID int64) ([]*domain.Product, error) {
	return []*domain.Product{m.product}, nil
}

func (m *productRepoMock) ListProductsByCategory(_ context.Context, categoryID int64) ([]*domain.Product, error) {
	return []*domain.Product{m.product}, nil
}

func (m *productRepoMock) ReserveProduct(_ context.Context, id int64, quantity int64) (*domain.Product, error) {
	m.reservedID = id
	m.reservedQuantity = quantity
	m.product.AvailableQuantity -= quantity
	return m.product, nil
}

func (m *productRepoMock) ReleaseProduct(_ context.Context, id int64, quantity int64) (*domain.Product, error) {
	m.product.AvailableQuantity += quantity
	return m.product, nil
}

func TestCreateProductDefaultsStatus(t *testing.T) {
	repo := &productRepoMock{}
	service := NewProductService(repo)

	id, err := service.CreateProduct(context.Background(), &dto.CreateProductRequest{
		SellerID:          1,
		CategoryID:        2,
		Name:              "Keyboard",
		Price:             19900,
		AvailableQuantity: 3,
	})
	if err != nil {
		t.Fatalf("CreateProduct returned error: %v", err)
	}

	if id != 10 {
		t.Fatalf("expected id 10, got %d", id)
	}
	if repo.product.Status != domain.StatusActive {
		t.Fatalf("expected status %s, got %s", domain.StatusActive, repo.product.Status)
	}
}

func TestReserveProductReturnsUpdatedStock(t *testing.T) {
	repo := &productRepoMock{
		product: &domain.Product{
			ID:                10,
			SellerID:          1,
			CategoryID:        2,
			Name:              "Keyboard",
			Price:             19900,
			AvailableQuantity: 5,
			Status:            domain.StatusActive,
		},
	}
	service := NewProductService(repo)

	resp, err := service.ReserveProduct(context.Background(), 10, &dto.ReserveProductRequest{Quantity: 2})
	if err != nil {
		t.Fatalf("ReserveProduct returned error: %v", err)
	}

	if repo.reservedID != 10 || repo.reservedQuantity != 2 {
		t.Fatalf("expected reserve call id=10 quantity=2, got id=%d quantity=%d", repo.reservedID, repo.reservedQuantity)
	}
	if resp.AvailableQuantity != 3 {
		t.Fatalf("expected stock 3, got %d", resp.AvailableQuantity)
	}
}
