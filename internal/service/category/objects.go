package category

type Category struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
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
