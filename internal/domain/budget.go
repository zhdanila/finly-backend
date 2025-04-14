package domain

import (
	"time"
)

type Budget struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Currency  string    `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
