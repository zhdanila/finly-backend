package category

import "time"

type Category struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name" validate:"required"`
	IsUserCategory bool      `json:"is_user_category"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	Category
}

type CreateCategoryResponse struct {
	Id string `json:"id"`
}

type GetCategoryByIDRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	ID     string `param:"id" validate:"required"`
}

type GetCategoryByIDResponse struct {
	Category
}

type ListCategoriesRequest struct {
	UserID string `header:"User-Id" validate:"required"`
}

type ListCategoriesResponse struct {
	Categories []Category `json:"categories"`
}

type DeleteCategoryRequest struct {
	UserID string `header:"User-Id" validate:"required"`
	ID     string `param:"id" validate:"required"`
}

type DeleteCategoryResponse struct{}
