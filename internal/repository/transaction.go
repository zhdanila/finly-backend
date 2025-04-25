package repository

import (
	"context"
	"encoding/json"
	"errors"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

const (
	TransactionTable = "transactions"

	TTL_ListTransactionsCache   = 5 * time.Minute
	TTL_GetByIDTransactionCache = 15 * time.Minute
)

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

func (t TransactionRepository) InvalidateCache(ctx context.Context, userID, transactionID string) error {
	keys := []string{
		fmt.Sprintf("transactions:user:%s", userID),
		fmt.Sprintf("transaction:%s:user:%s", transactionID, userID),
	}
	if err := t.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
		return err
	}
	zap.L().Sugar().Infof("Cache invalidated, userID: %s, transactionID: %s", userID, transactionID)
	return nil
}

func (t TransactionRepository) GetDB() *sqlx.DB {
	return t.postgres
}

func (t TransactionRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, userID string, budgetID string, categoryID string, transactionType string, note string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, budget_id, category_id, amount, transaction_type, note) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", TransactionTable)
	var transactionID string
	if err := tx.QueryRowContext(ctx, query, userID, budgetID, categoryID, amount, transactionType, note).Scan(&transactionID); err != nil {
		return "", err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	return transactionID, nil
}

func (t TransactionRepository) List(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	cacheKey := fmt.Sprintf("transactions:user:%s", userID)

	var transactions []*domain.Transaction
	cachedTransactions, err := t.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedTransactions), &transactions); err == nil {
			zap.L().Sugar().Infof("Cache hit for transactions, key: %s", cacheKey)
			return transactions, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached transactions, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, userID: %s, error: %v", cacheKey, userID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for transactions, key: %s", cacheKey)
	}

	query := "SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC"
	if err = t.postgres.SelectContext(ctx, &transactions, query, userID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch transactions from DB, userID: %s, error: %v", userID, err)
		return nil, err
	}

	serializedTransactions, err := json.Marshal(transactions)
	if err == nil {
		if err = t.redis.Set(ctx, cacheKey, serializedTransactions, TTL_ListTransactionsCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache transactions, key: %s, userID: %s, error: %v", cacheKey, userID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached transactions, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal transactions for caching, userID: %s, error: %v", userID, err)
	}

	return transactions, nil
}

func (t TransactionRepository) UpdateTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string, categoryID, transactionType, note string, amount float64) error {
	query := fmt.Sprintf("UPDATE %s SET category_id = $1, transaction_type = $2, note = $3, amount = $4 WHERE id = $5 AND user_id = $6", TransactionTable)
	if _, err := tx.ExecContext(ctx, query, categoryID, transactionType, note, amount, transactionID, userID); err != nil {
		return err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after update, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	return nil
}

func (t TransactionRepository) DeleteTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)
	_, err := tx.ExecContext(ctx, query, transactionID, userID)
	if err != nil {
		return err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after delete, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	return nil
}

func (t TransactionRepository) GetByID(ctx context.Context, transactionID, userID string) (*domain.Transaction, error) {
	cacheKey := fmt.Sprintf("transaction:%s:user:%s", transactionID, userID)

	var transaction domain.Transaction
	cachedTransaction, err := t.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedTransaction), &transaction); err == nil {
			zap.L().Sugar().Infof("Cache hit for transaction, key: %s", cacheKey)
			return &transaction, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached transaction, key: %s, error: %v", cacheKey, err)
	} else if err != redis.Nil {
		zap.L().Sugar().Errorf("Redis error for key: %s, userID: %s, transactionID: %s, error: %v", cacheKey, userID, transactionID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for transaction, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)
	if err = t.postgres.GetContext(ctx, &transaction, query, transactionID, userID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch transaction from DB, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
		return nil, err
	}

	serializedTransaction, err := json.Marshal(transaction)
	if err == nil {
		if err = t.redis.Set(ctx, cacheKey, serializedTransaction, TTL_GetByIDTransactionCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache transaction, key: %s, userID: %s, transactionID: %s, error: %v", cacheKey, userID, transactionID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached transaction, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal transaction for caching, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	return &transaction, nil
}
