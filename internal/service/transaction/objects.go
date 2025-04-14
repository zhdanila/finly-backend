package transaction

import (
	"finly-backend/internal/domain/enums/e_transaction_type"
	"time"
)

type Transaction struct {
	ID         string                  `json:"id"`
	UserID     string                  `json:"user_id"`
	BudgetID   string                  `json:"budget_id"`
	CategoryID string                  `json:"category_id"`
	Amount     int64                   `json:"amount"`
	Type       e_transaction_type.Enum `json:"type"`
	Note       string                  `json:"note"`
	CreatedAt  time.Time               `json:"created_at"`
}

type CreateTransactionRequest struct {
	UserID     string                  `header:"User-Id" validate:"required"`
	CategoryID string                  `json:"category_id" validate:"required"`
	BudgetID   string                  `json:"budget_id" validate:"required"`
	Amount     float64                 `json:"amount" validate:"required"`
	Type       e_transaction_type.Enum `json:"type" validate:"required,oneof=deposit withdrawal"`
	Note       string                  `json:"note"`
}

type CreateTransactionResponse struct {
	ID string `json:"id"`
}
