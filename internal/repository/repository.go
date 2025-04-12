package repository

import (
	"context"
	"finly-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Auth
}

func NewRepository(postgres *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Auth: NewAuthRepository(postgres, redis),
	}
}

type Auth interface {
	Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)

	AddTokenToBlacklist(ctx context.Context, token string, ttlSeconds float64) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	RemoveToken(ctx context.Context, token string) error
}
