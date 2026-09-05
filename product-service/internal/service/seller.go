package service

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/repository"
)

type SellerService struct {
	sellerRepo repository.SellerRepository
}

func NewSellerService(sellerRepo repository.SellerRepository) *SellerService {
	return &SellerService{
		sellerRepo: sellerRepo,
	}
}

func (s *SellerService) CreateSeller(ctx context.Context, req *dto.CreateSellerRequest) (*dto.GetSellerByIDResponse, error) {
	seller := &domain.Seller{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	err := s.sellerRepo.CreateSeller(ctx, seller)
	if err != nil {
		return nil, err
	}

	return sellerToResponse(seller), nil
}

func (s *SellerService) GetSellerByID(ctx context.Context, req *dto.GetSellerByIDRequest) (*dto.GetSellerByIDResponse, error) {
	seller, err := s.sellerRepo.GetSellerByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return sellerToResponse(seller), nil
}

func sellerToResponse(seller *domain.Seller) *dto.GetSellerByIDResponse {
	return &dto.GetSellerByIDResponse{
		ID:        seller.ID,
		FirstName: seller.FirstName,
		LastName:  seller.LastName,
		Email:     seller.Email,
		CreatedAt: seller.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: seller.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *SellerService) UpdateSeller(ctx context.Context, req *dto.UpdateSellerRequest) error {
	domainSeller := &domain.Seller{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	err := s.sellerRepo.UpdateSeller(ctx, domainSeller)
	if err != nil {
		return err
	}

	return nil
}

func (s *SellerService) DeleteSeller(ctx context.Context, req *dto.DeleteSellerRequest) error {
	err := s.sellerRepo.DeleteSeller(ctx, req.ID)
	if err != nil {
		return err
	}

	return nil

}
