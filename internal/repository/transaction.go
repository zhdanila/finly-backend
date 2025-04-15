package repository

import (
	"context"
	"errors"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strings"
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
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", TransactionTable)

	var transactions []*domain.Transaction
	if err := t.postgres.SelectContext(ctx, &transactions, query, userID); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t TransactionRepository) Update(ctx context.Context, transactionID, userID string, categoryID, transactionType, note *string, amount *float64) error {
	var (
		setParts []string
		args     []interface{}
		argIndex = 1
	)

	if categoryID != nil {
		setParts = append(setParts, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *categoryID)
		argIndex++
	}
	if transactionType != nil {
		setParts = append(setParts, fmt.Sprintf("transaction_type = $%d", argIndex))
		args = append(args, *transactionType)
		argIndex++
	}
	if note != nil {
		setParts = append(setParts, fmt.Sprintf("note = $%d", argIndex))
		args = append(args, *note)
		argIndex++
	}
	if amount != nil {
		setParts = append(setParts, fmt.Sprintf("amount = $%d", argIndex))
		args = append(args, *amount)
		argIndex++
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	args = append(args, transactionID, userID)
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d AND user_id = $%d",
		TransactionTable,
		strings.Join(setParts, ", "),
		argIndex, argIndex+1,
	)

	_, err := t.postgres.ExecContext(ctx, query, args...)
	return err
}
