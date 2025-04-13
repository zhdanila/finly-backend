package repository

import (
	"context"
	"finly-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Auth
	Budget
}

func NewRepository(postgres *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Auth:   NewAuthRepository(postgres, redis),
		Budget: NewBudgetRepository(postgres, redis),
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

type Budget interface {
	Create(ctx context.Context, userID, name, currency string) error
	GetByID(ctx context.Context, budgetID, userID string) (*domain.Budget, error)
	List(ctx context.Context, userID string) ([]*domain.Budget, error)
	Delete(ctx context.Context, budgetID, userID string) error
}
