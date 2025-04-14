package budget

import "time"

type Budget struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateBudgetRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type CreateBudgetResponse struct {
	ID string `json:"id"`
}

type GetBudgetByIDRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	BudgetID string `param:"budget_id" validate:"required,uuid"`
}

type GetBudgetByIDResponse struct {
	Budget
}
