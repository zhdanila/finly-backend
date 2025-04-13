package repository

import (
	"context"
	"finly-backend/internal/domain"
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

func (b BudgetRepository) GetByID(ctx context.Context, budgetID, userID string) (*domain.Budget, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", BudgetTable)

	var budget domain.Budget
	if err := b.postgres.GetContext(ctx, &budget, query, budgetID, userID); err != nil {
		return nil, err
	}

	return &budget, nil
}

func (b BudgetRepository) List(ctx context.Context, userID string) ([]*domain.Budget, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", BudgetTable)

	var budgets []*domain.Budget
	if err := b.postgres.SelectContext(ctx, &budgets, query, userID); err != nil {
		return nil, err
	}

	return budgets, nil
}
