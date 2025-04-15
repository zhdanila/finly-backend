package repository

import (
	"context"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"time"
)

const BudgetHistoryTable = "budgets_history"

type BudgetHistoryRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func (b BudgetHistoryRepository) CreateInitialTX(ctx context.Context, tx *sqlx.Tx, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func NewBudgetHistoryRepository(postgres *sqlx.DB, redis *redis.Client) *BudgetHistoryRepository {
	return &BudgetHistoryRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (b BudgetHistoryRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, budgetID, transactionID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance, transaction_id) VALUES ($1, $2, $3) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount, transactionID).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (b BudgetHistoryRepository) GetLastByBudgetID(ctx context.Context, budgetID string) (*domain.BudgetHistory, error) {
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

func (b BudgetHistoryRepository) ListFromDate(ctx context.Context, budgetID string, fromDate time.Time) ([]*domain.BudgetHistory, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 AND created_at > $2 ORDER BY created_at ASC", BudgetHistoryTable)

	var histories []*domain.BudgetHistory
	if err := b.postgres.SelectContext(ctx, &histories, query, budgetID, fromDate); err != nil {
		return nil, err
	}

	return histories, nil
}

func (b BudgetHistoryRepository) UpdateBalanceTX(ctx context.Context, tx *sqlx.Tx, transactionID string, amount float64) error {
	query := fmt.Sprintf("UPDATE %s SET balance = $1 WHERE transaction_id = $2", BudgetHistoryTable)

	if _, err := tx.ExecContext(ctx, query, amount, transactionID); err != nil {
		return err
	}

	return nil
}

func (b BudgetHistoryRepository) GetCurrentBalance(ctx context.Context, budgetID string) (float64, error) {
	query := fmt.Sprintf("SELECT balance FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)

	var balance float64
	if err := b.postgres.GetContext(ctx, &balance, query, budgetID); err != nil {
		return 0, err
	}

	return balance, nil
}
