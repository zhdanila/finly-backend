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
	Category
}

func NewRepository(postgres *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Auth:     NewAuthRepository(postgres, redis),
		Budget:   NewBudgetRepository(postgres, redis),
		Category: NewCategoryRepository(postgres, redis),
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
	Create(ctx context.Context, userID, currency string) (string, error)
	GetByID(ctx context.Context, budgetID, userID string) (*domain.Budget, error)
}

type Category interface {
	Create(ctx context.Context, userID, name, description string) (string, error)
	GetByID(ctx context.Context, categoryID, userID string) (*domain.Category, error)
	List(ctx context.Context, userID string) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID, userID string) error
}
