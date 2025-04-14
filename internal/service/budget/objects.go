package budget

import (
	"time"
)

type Budget struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateBudgetRequest struct {
	UserID   string  `header:"User-Id" validate:"required"`
	Currency string  `json:"currency" validate:"required"`
	Amount   float64 `json:"amount" validate:"required"`
}

type CreateBudgetResponse struct {
	ID string `json:"id"`
}

type GetBudgetByIDRequest struct {
	UserID string `header:"User-Id" validate:"required"`
}

type GetBudgetByIDResponse struct {
	*Budget
}

type GetBudgetHistoryRequest struct {
	BudgetID string `param:"budget_id" validate:"required"`
}

type GetBudgetHistoryResponse struct {
	BudgetHistory []*BudgetHistory `json:"budget_history"`
}

type BudgetHistory struct {
	ID        string    `json:"id"`
	BudgetID  string    `json:"budget_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}
