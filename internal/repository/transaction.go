package repository

import (
	"context"
	"finly-backend/internal/domain"
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

func (t TransactionRepository) List(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1 ORDER BY created_at DESC", TransactionTable)

	var transactions []*domain.Transaction
	if err := t.postgres.SelectContext(ctx, &transactions, query, userID); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t TransactionRepository) UpdateTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string, categoryID, transactionType, note string, amount float64) error {
	query := fmt.Sprintf("UPDATE %s SET category_id = $1, transaction_type = $2, note = $3, amount = $4 WHERE id = $5 AND user_id = $6", TransactionTable)
	if _, err := tx.ExecContext(ctx, query, categoryID, transactionType, note, amount, transactionID, userID); err != nil {
		return err
	}

	return nil
}

func (t TransactionRepository) DeleteTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)
	_, err := tx.ExecContext(ctx, query, transactionID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (t TransactionRepository) GetByID(ctx context.Context, transactionID, userID string) (*domain.Transaction, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)

	var transaction domain.Transaction
	if err := t.postgres.GetContext(ctx, &transaction, query, transactionID, userID); err != nil {
		return nil, err
	}

	return &transaction, nil
}
