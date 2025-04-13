package domain

import (
	"time"
)

type Category struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
