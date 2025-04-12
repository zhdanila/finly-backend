package repository

import (
	"context"
	"finly-backend/internal/domain"
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
	Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}
