package domain

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Budget struct {
	ID        uuid.UUID     `db:"id"`
	UserID    uuid.UUID     `db:"user_id"`
	Name      string        `db:"name"`
	Amount    sql.NullInt64 `db:"amount"`
	Currency  string        `db:"currency"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}
