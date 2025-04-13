package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const BudgetTable = "budget"

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

func (b BudgetRepository) Create(ctx context.Context, userId, name, currency string) error {
	query := fmt.Sprintf("INSERT INTO %s (user_id, name, currency) VALUES ($1, $2, $3)", BudgetTable)

	if _, err := b.postgres.ExecContext(ctx, query, userId, name, currency); err != nil {
		return err
	}

	return nil
}
