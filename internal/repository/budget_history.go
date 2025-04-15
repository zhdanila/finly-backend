package repository

import (
	"context"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const BudgetHistoryTable = "budgets_history"

type BudgetHistoryRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewBudgetHistoryRepository(postgres *sqlx.DB, redis *redis.Client) *BudgetHistoryRepository {
	return &BudgetHistoryRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (b BudgetHistoryRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (b BudgetHistoryRepository) GetLastByID(ctx context.Context, budgetID string) (*domain.BudgetHistory, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)

	var history domain.BudgetHistory
	if err := b.postgres.GetContext(ctx, &history, query, budgetID); err != nil {
		return nil, err
	}

	return &history, nil
}

func (b BudgetHistoryRepository) Create(ctx context.Context, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := b.postgres.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (b BudgetHistoryRepository) List(ctx context.Context, budgetID string) ([]*domain.BudgetHistory, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at ASC", BudgetHistoryTable)

	var histories []*domain.BudgetHistory
	if err := b.postgres.SelectContext(ctx, &histories, query, budgetID); err != nil {
		return nil, err
	}

	return histories, nil
}
