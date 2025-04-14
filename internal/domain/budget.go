package domain

import (
	"database/sql"
	"time"
)

type Budget struct {
	ID        string        `db:"id"`
	UserID    string        `db:"user_id"`
	Amount    sql.NullInt64 `db:"amount"`
	Currency  string        `db:"currency"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}
