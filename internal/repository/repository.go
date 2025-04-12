package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
)

const UsersTable = "users"

type Repository struct {
	Auth
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Auth: NewAuthRepository(db),
	}
}

type Auth interface {
	Register(ctx context.Context, email, passwordHash, firstName, lastName string) error
}
