package domain

import (
	"database/sql"
	"time"
)

type BudgetHistory struct {
	ID            string         `db:"id"`
	TransactionID sql.NullString `db:"transaction_id"`
	BudgetID      string         `db:"budget_id"`
	Balance       float64        `db:"balance"`
	CreatedAt     time.Time      `db:"created_at"`
}
