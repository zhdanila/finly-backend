package repository

import (
	"context"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strings"
)

const TransactionTable = "transactions"

type TransactionRepository struct {
	postgres          *sqlx.DB
	redis             *redis.Client
	bidgetHistoryRepo *BudgetHistoryRepository
}

func NewTransactionRepository(postgres *sqlx.DB, redis *redis.Client, budgetHistoryRepo *BudgetHistoryRepository) *TransactionRepository {
	return &TransactionRepository{
		postgres:          postgres,
		redis:             redis,
		bidgetHistoryRepo: budgetHistoryRepo,
	}
}

func (t TransactionRepository) Create(ctx context.Context, userID string, budgetID string, categoryID string, transactionType string, note string, amount float64) (string, error) {
	tx, err := t.postgres.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := fmt.Sprintf("INSERT INTO %s (user_id, budget_id, category_id, amount, transaction_type, note) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", TransactionTable)
	var transactionID string
	if err := tx.QueryRowContext(ctx, query, userID, budgetID, categoryID, amount, transactionType, note).Scan(&transactionID); err != nil {
		return "", err
	}

	budget, err := t.bidgetHistoryRepo.GetLastByID(ctx, budgetID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			budget = nil
		} else {
			return "", err
		}
	}

	var newAmount float64
	switch transactionType {
	case e_transaction_type.Deposit.String():
		if budget == nil {
			newAmount = amount
		} else {
			newAmount = budget.Balance + amount
		}
	case e_transaction_type.Withdrawal.String():
		if budget == nil || budget.Balance < amount {
			return "", fmt.Errorf("not enough budget")
		}
		newAmount = budget.Balance - amount
	default:
		return "", fmt.Errorf("invalid transaction type")
	}

	if _, err = t.bidgetHistoryRepo.CreateTX(ctx, tx, budgetID, newAmount); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return transactionID, nil
}
