package category

import (
	"finly-backend/internal/domain"
	"time"
)

type CategoryObject struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name" validate:"required"`
	IsUserCategory bool      `json:"is_user_category"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	CategoryObject
}

type CreateCategoryResponse struct {
	Id string `json:"id"`
}

type GetCategoryByIDRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	ID     string `param:"id" validate:"required"`
}

type GetCategoryByIDResponse struct {
	*CategoryObject
}

type ListCategoriesRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	Test   string `query:"test"`
}

type ListCategoriesResponse struct {
	Categories []CategoryObject `json:"categories"`
}

type DeleteCategoryRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	ID     string `param:"id" validate:"required"`
}

type DeleteCategoryResponse struct{}

type ListCustomCategoriesRequest struct {
	UserID string `header:"User-Id" validate:"required"`
}

type ListCustomCategoriesResponse struct {
	Categories []CategoryObject `json:"categories"`
}

func convertCategories(categories []*domain.Category) []CategoryObject {
	categoriesResponse := make([]CategoryObject, len(categories))
	for i, category := range categories {
		categoriesResponse[i] = CategoryObject{
			ID:             category.ID,
			UserID:         category.UserID.String,
			Name:           category.Name,
			IsUserCategory: category.IsUserCategory,
			CreatedAt:      category.CreatedAt,
		}
	}
	return categoriesResponse
}
