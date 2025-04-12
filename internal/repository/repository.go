package repository

import (
	"context"
	"finly-backend/internal/domain"
	"finly-backend/internal/repository/postgres"
	redis2 "finly-backend/internal/repository/redis"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Auth
	TokenBlacklist
}

func NewRepository(db *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Auth:           postgres.NewAuthRepository(db),
		TokenBlacklist: redis2.NewTokenBlacklistRepository(redis),
	}
}

type Auth interface {
	Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type TokenBlacklist interface {
	AddToken(ctx context.Context, token string, ttlSeconds float64) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
