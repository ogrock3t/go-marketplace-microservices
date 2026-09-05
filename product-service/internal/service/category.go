package service

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/repository"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.GetCategoryByIDResponse, error) {
	category := &domain.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Description: req.Description,
	}

	err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	return &dto.GetCategoryByIDResponse{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Description: category.Description,
	}, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*domain.Category, error) {
	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.Category{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Description: category.Description,
	}, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *dto.UpdateCategoryRequest) error {
	categoryToUpdate := &domain.Category{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Description: category.Description,
	}

	err := s.categoryRepo.UpdateCategory(ctx, categoryToUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	err := s.categoryRepo.DeleteCategory(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context) (*dto.ListCategoriesResponse, error) {
	list, err := s.categoryRepo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]*dto.GetCategoryByIDResponse, 0, len(list))
	for _, category := range list {
		categories = append(categories, &dto.GetCategoryByIDResponse{
			ID:          category.ID,
			ParentID:    category.ParentID,
			Name:        category.Name,
			Description: category.Description,
		})
	}

	return &dto.ListCategoriesResponse{
		Categories: categories,
	}, nil
}

func (s *CategoryService) ListSubcategories(ctx context.Context, id int64) (*dto.ListSubcategoriesResponse, error) {
	list, err := s.categoryRepo.ListSubcategories(ctx, id)
	if err != nil {
		return nil, err
	}

	categories := make([]*dto.GetCategoryByIDResponse, 0, len(list))
	for _, category := range list {
		categories = append(categories, &dto.GetCategoryByIDResponse{
			ID:          category.ID,
			ParentID:    category.ParentID,
			Name:        category.Name,
			Description: category.Description,
		})
	}

	return &dto.ListSubcategoriesResponse{
		Categories: categories,
	}, nil
}
