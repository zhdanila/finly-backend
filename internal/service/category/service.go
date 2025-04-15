package category

import (
	"context"
	"finly-backend/internal/repository"
)

type Service struct {
	repo repository.Category
}

func NewService(repo repository.Category) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateCategoryRequest) (*CreateCategoryResponse, error) {
	var err error

	id, err := s.repo.Create(ctx, req.UserID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return &CreateCategoryResponse{
		Id: id,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, req *GetCategoryByIDRequest) (*GetCategoryByIDResponse, error) {
	var err error

	category, err := s.repo.GetByID(ctx, req.ID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetCategoryByIDResponse{
		Category{
			ID:             category.ID,
			UserID:         category.UserID.String,
			Name:           category.Name,
			Description:    category.Description,
			IsUserCategory: category.IsUserCategory,
			CreatedAt:      category.CreatedAt,
		},
	}, nil
}

func (s *Service) List(ctx context.Context, req *ListCategoriesRequest) (*ListCategoriesResponse, error) {
	var err error

	categories, err := s.repo.List(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	categoriesResponse := make([]Category, len(categories))
	for i, category := range categories {
		categoriesResponse[i] = Category{
			ID:             category.ID,
			UserID:         category.UserID.String,
			Name:           category.Name,
			Description:    category.Description,
			IsUserCategory: category.IsUserCategory,
			CreatedAt:      category.CreatedAt,
		}
	}

	return &ListCategoriesResponse{
		Categories: categoriesResponse,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *DeleteCategoryRequest) (*DeleteCategoryResponse, error) {
	var err error

	if err = s.repo.Delete(ctx, req.ID, req.UserID); err != nil {
		return nil, err
	}

	return &DeleteCategoryResponse{}, nil
}
