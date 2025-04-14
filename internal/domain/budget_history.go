package domain

import (
	"time"
)

type BudgetHistory struct {
	ID        string    `db:"id"`
	BudgetID  string    `db:"budget_id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
}
