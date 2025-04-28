package transaction

import (
	"context"
	"finly-backend/internal/domain"
	"finly-backend/pkg/db"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type Transaction interface {
	CreateTX(ctx context.Context, tx *sqlx.Tx, userID, budgetID, categoryID, transactionType, note string, amount float64) (string, error)
	GetDB() *sqlx.DB
	List(ctx context.Context, userID string) ([]*domain.Transaction, error)
	UpdateTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string, categoryID, transactionType, note string, amount float64) error
	DeleteTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string) error
	GetByID(ctx context.Context, transactionID, userID string) (*domain.Transaction, error)
}

const (
	TransactionTable = "transactions"

	TTL_ListTransactionsCache   = 5 * time.Minute
	TTL_GetByIDTransactionCache = 15 * time.Minute

	cacheKeyTransactionByIDAndUser = "transaction:%s:user:%s"
	cacheKeyTransactionsByUser     = "transactions:user:%s"
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

func (t *TransactionRepository) cacheKeys(userID, transactionID string) []string {
	return []string{
		fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID),
		fmt.Sprintf(cacheKeyTransactionsByUser, userID),
	}
}

func (t *TransactionRepository) InvalidateCache(ctx context.Context, userID, transactionID string) error {
	zap.L().Sugar().Infof("Invalidating transaction cache for userID: %s, transactionID: %s", userID, transactionID)

	keys := t.cacheKeys(userID, transactionID)
	if err := t.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate transaction cache for userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
		return err
	}

	zap.L().Sugar().Infof("Transaction cache invalidated for userID: %s, transactionID: %s", userID, transactionID)
	return nil
}

func (t *TransactionRepository) GetDB() *sqlx.DB {
	return t.postgres
}

func (t *TransactionRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, userID string, budgetID string, categoryID string, transactionType string, note string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, budget_id, category_id, amount, transaction_type, note) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", TransactionTable)
	var transactionID string
	if err := tx.QueryRowContext(ctx, query, userID, budgetID, categoryID, amount, transactionType, note).Scan(&transactionID); err != nil {
		zap.L().Sugar().Errorf("Error creating transaction, userID: %s, error: %v", userID, err)
		return "", err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	zap.L().Sugar().Infof("Transaction created successfully, transactionID: %s", transactionID)
	return transactionID, nil
}

func (t *TransactionRepository) List(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	cacheKey := fmt.Sprintf(cacheKeyTransactionsByUser, userID)
	fetch := func() ([]*domain.Transaction, error) {
		var transactions []*domain.Transaction
		query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1 ORDER BY created_at DESC", TransactionTable)
		if err := t.postgres.SelectContext(ctx, &transactions, query, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch transactions from DB for userID: %s, error: %v", userID, err)
			return nil, err
		}
		return transactions, nil
	}

	return db.WithCache(ctx, t.redis, cacheKey, TTL_ListTransactionsCache, fetch)
}

func (t *TransactionRepository) UpdateTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string, categoryID, transactionType, note string, amount float64) error {
	query := fmt.Sprintf("UPDATE %s SET category_id = $1, transaction_type = $2, note = $3, amount = $4 WHERE id = $5 AND user_id = $6", TransactionTable)
	if _, err := tx.ExecContext(ctx, query, categoryID, transactionType, note, amount, transactionID, userID); err != nil {
		zap.L().Sugar().Errorf("Error updating transaction, transactionID: %s, userID: %s, error: %v", transactionID, userID, err)
		return err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after update, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	zap.L().Sugar().Infof("Transaction updated successfully, transactionID: %s, userID: %s", transactionID, userID)
	return nil
}

func (t *TransactionRepository) DeleteTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)
	_, err := tx.ExecContext(ctx, query, transactionID, userID)
	if err != nil {
		zap.L().Sugar().Errorf("Error deleting transaction, transactionID: %s, userID: %s, error: %v", transactionID, userID, err)
		return err
	}

	if err := t.InvalidateCache(ctx, userID, transactionID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after delete, userID: %s, transactionID: %s, error: %v", userID, transactionID, err)
	}

	zap.L().Sugar().Infof("Transaction deleted successfully, transactionID: %s, userID: %s", transactionID, userID)
	return nil
}

func (t *TransactionRepository) GetByID(ctx context.Context, transactionID, userID string) (*domain.Transaction, error) {
	cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)
	fetch := func() (*domain.Transaction, error) {
		var transaction domain.Transaction
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", TransactionTable)
		if err := t.postgres.GetContext(ctx, &transaction, query, transactionID, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch transaction from DB, transactionID: %s, userID: %s, error: %v", transactionID, userID, err)
			return nil, err
		}
		return &transaction, nil
	}

	return db.WithCache(ctx, t.redis, cacheKey, TTL_GetByIDTransactionCache, fetch)
}
