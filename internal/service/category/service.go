package category

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/repository/category"
	"go.uber.org/zap"
)

type Category interface {
	Create(ctx context.Context, req *CreateCategoryRequest) (*CreateCategoryResponse, error)
	GetByID(ctx context.Context, req *GetCategoryByIDRequest) (*GetCategoryByIDResponse, error)
	List(ctx context.Context, req *ListCategoriesRequest) (*ListCategoriesResponse, error)
	ListCustom(ctx context.Context, req *ListCustomCategoriesRequest) (*ListCustomCategoriesResponse, error)
	Delete(ctx context.Context, req *DeleteCategoryRequest) (*DeleteCategoryResponse, error)
}

type Service struct {
	repo category.Category
}

func NewService(repo category.Category) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateCategoryRequest) (*CreateCategoryResponse, error) {
	id, err := s.repo.Create(ctx, req.UserID, req.Name)
	if err != nil {
		zap.L().Sugar().Errorf("Create: failed for userID=%s, categoryName=%s: %v", req.UserID, req.Name, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Create: category created with id=%s for userID=%s", id, req.UserID)
	return &CreateCategoryResponse{
		Id: id,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, req *GetCategoryByIDRequest) (*GetCategoryByIDResponse, error) {
	category, err := s.repo.GetByID(ctx, req.ID, req.UserID)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			zap.L().Sugar().Infof("GetByID: no category found for categoryID=%s, userID=%s", req.ID, req.UserID)
			return &GetCategoryByIDResponse{}, nil
		}
		zap.L().Sugar().Errorf("GetByID: failed for categoryID=%s, userID=%s: %v", req.ID, req.UserID, err)
		return nil, err
	}

	return &GetCategoryByIDResponse{
		&CategoryObject{
			ID:             category.ID,
			UserID:         category.UserID.String,
			Name:           category.Name,
			IsUserCategory: category.IsUserCategory,
			CreatedAt:      category.CreatedAt,
		},
	}, nil
}

func (s *Service) List(ctx context.Context, req *ListCategoriesRequest) (*ListCategoriesResponse, error) {
	categories, err := s.repo.List(ctx, req.UserID)
	if err != nil {
		zap.L().Sugar().Errorf("List: failed for userID=%s: %v", req.UserID, err)
		return nil, err
	}

	categoriesResponse := convertCategories(categories)

	return &ListCategoriesResponse{
		Categories: categoriesResponse,
	}, nil
}

func (s *Service) ListCustom(ctx context.Context, req *ListCustomCategoriesRequest) (*ListCustomCategoriesResponse, error) {
	categories, err := s.repo.ListCustom(ctx, req.UserID)
	if err != nil {
		zap.L().Sugar().Errorf("ListCustom: failed for userID=%s: %v", req.UserID, err)
		return nil, err
	}

	categoriesResponse := convertCategories(categories)

	zap.L().Sugar().Infof("ListCustom: found %d custom categories for userID=%s", len(categoriesResponse), req.UserID)
	return &ListCustomCategoriesResponse{
		Categories: categoriesResponse,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *DeleteCategoryRequest) (*DeleteCategoryResponse, error) {
	if err := s.repo.Delete(ctx, req.ID, req.UserID); err != nil {
		zap.L().Sugar().Errorf("Delete: failed to delete categoryID=%s, userID=%s: %v", req.ID, req.UserID, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Delete: successfully deleted categoryID=%s for userID=%s", req.ID, req.UserID)
	return &DeleteCategoryResponse{}, nil
}
