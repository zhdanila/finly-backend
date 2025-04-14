package repository

import (
	"context"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const BudgetTable = "budgets"

type BudgetRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewBudgetRepository(postgres *sqlx.DB, redis *redis.Client) *BudgetRepository {
	return &BudgetRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (b BudgetRepository) Create(ctx context.Context, userId, currency string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, currency) VALUES ($1, $2) RETURNING id", BudgetTable)

	var id string
	if err := b.postgres.QueryRowContext(ctx, query, userId, currency).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (b BudgetRepository) GetByUserID(ctx context.Context, userID string) (*domain.Budget, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", BudgetTable)

	var budget domain.Budget
	if err := b.postgres.GetContext(ctx, &budget, query, userID); err != nil {
		return nil, err
	}

	return &budget, nil
}
