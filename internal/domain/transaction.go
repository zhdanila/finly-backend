package domain

import "time"

type Transaction struct {
	ID              string    `db:"id"`
	UserID          string    `db:"user_id"`
	BudgetID        string    `db:"budget_id"`
	CategoryID      string    `db:"category_id"`
	Amount          float64   `db:"amount"`
	TransactionType string    `db:"transaction_type"`
	Note            string    `db:"note"`
	CreatedAt       time.Time `db:"created_at"`
}
