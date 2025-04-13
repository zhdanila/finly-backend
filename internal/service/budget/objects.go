package budget

type Budget struct {
	Name     string `json:"name"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type CreateBudgetRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type CreateBudgetResponse struct{}
