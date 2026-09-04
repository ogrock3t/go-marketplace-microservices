package service

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/repository"
)

type ProductService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (int64, error) {
	product := &domain.Product{
		SellerID:          req.SellerID,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		AvailableQuantity: req.AvailableQuantity,
	}

	if req.Status == "" {
		if req.AvailableQuantity == 0 {
			product.Status = domain.StatusOutOfStock
		} else {
			product.Status = domain.StatusActive
		}
	} else {
		product.SetStatus(req.Status)
	}

	if err := s.productRepo.CreateProduct(ctx, product); err != nil {
		return 0, err
	}

	return product.ID, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*dto.GetProductByIDResponse, error) {
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return productToResponse(product), nil
}

func productToResponse(product *domain.Product) *dto.GetProductByIDResponse {
	return &dto.GetProductByIDResponse{
		ID:                product.ID,
		SellerID:          product.SellerID,
		CategoryID:        product.CategoryID,
		Name:              product.Name,
		Description:       product.Description,
		Price:             product.Price,
		AvailableQuantity: product.AvailableQuantity,
		Status:            product.GetStatus(),
	}
}

func (s *ProductService) UpdateProductByID(ctx context.Context, req *dto.UpdateProductByIDRequest) error {
	product := &domain.Product{
		ID:                req.ID,
		SellerID:          req.SellerID,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		AvailableQuantity: req.AvailableQuantity,
	}

	product.SetStatus(req.Status)

	if err := s.productRepo.UpdateProductByID(ctx, product); err != nil {
		return err
	}

	return nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	return s.productRepo.DeleteProduct(ctx, id)
}

func (s *ProductService) ListProductsBySeller(ctx context.Context, id int64) (*dto.ListProductsBySellerResponse, error) {
	list, err := s.productRepo.ListProductsBySeller(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ListProductsBySellerResponse{
		Products: func() []*dto.GetProductByIDResponse {
			res := make([]*dto.GetProductByIDResponse, len(list))
			for i, product := range list {
				res[i] = productToResponse(product)
			}
			return res
		}(),
	}, nil
}

func (s *ProductService) ListProductsByCategory(ctx context.Context, id int64) (*dto.ListProductsByCategoryResponse, error) {
	list, err := s.productRepo.ListProductsByCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ListProductsByCategoryResponse{
		Products: func() []*dto.GetProductByIDResponse {
			res := make([]*dto.GetProductByIDResponse, len(list))
			for i, product := range list {
				res[i] = productToResponse(product)
			}
			return res
		}(),
	}, nil
}

func (s *ProductService) ReserveProduct(ctx context.Context, id int64, req *dto.ReserveProductRequest) (*dto.GetProductByIDResponse, error) {
	product, err := s.productRepo.ReserveProduct(ctx, id, req.Quantity)
	if err != nil {
		return nil, err
	}

	return productToResponse(product), nil
}
