package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const TransactionTable = "transactions"

type TransactionRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewTransactionRepository(postgres *sqlx.DB, redis *redis.Client) *TransactionRepository {
	return &TransactionRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (t TransactionRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, userID string, budgetID string, categoryID string, transactionType string, note string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, budget_id, category_id, amount, transaction_type, note) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", TransactionTable)
	var transactionID string
	if err := tx.QueryRowContext(ctx, query, userID, budgetID, categoryID, amount, transactionType, note).Scan(&transactionID); err != nil {
		return "", err
	}

	return transactionID, nil
}

func (t TransactionRepository) GetDB() *sqlx.DB {
	return t.postgres
}
