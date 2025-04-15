package domain

import (
	"database/sql"
	"time"
)

type Category struct {
	ID             string         `db:"id"`
	UserID         sql.NullString `db:"user_id"`
	Name           string         `db:"name"`
	Description    string         `db:"description"`
	IsUserCategory bool           `db:"is_user_category"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
