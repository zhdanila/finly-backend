package budget_history

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

const (
	BudgetHistoryTable = "budgets_history"

	TTL_GetLastHistoryByBudgetIDCache  = 5 * time.Minute
	TTL_ListBudgetHistoryCache         = 15 * time.Minute
	TTL_ListBudgetHistoryFromDateCache = 15 * time.Minute
	TTL_GetCurrentBalanceCache         = 30 * time.Second

	cacheKeyLastHistory = "budget:history:last:%s"
	cacheKeyListHistory = "budget:history:list:%s"
	cacheKeyBalance     = "budget:balance:%s"
	cacheKeyHistoryFrom = "budget:history:from:%s:date:%d:inclusive:%t"
)

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

func (b BudgetHistoryRepository) cacheKeys(budgetID string, fromDate ...time.Time) []string {
	keys := []string{
		fmt.Sprintf(cacheKeyLastHistory, budgetID),
		fmt.Sprintf(cacheKeyListHistory, budgetID),
		fmt.Sprintf(cacheKeyBalance, budgetID),
	}
	if len(fromDate) > 0 {
		keys = append(keys, fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate[0].Unix(), true))
		keys = append(keys, fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate[0].Unix(), false))
	}
	return keys
}

func (b BudgetHistoryRepository) InvalidateCache(ctx context.Context, budgetID string) error {
	zap.L().Sugar().Infof("Invalidating cache for budgetID: %s", budgetID)

	keys := b.cacheKeys(budgetID)
	if err := b.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache for budgetID: %s, error: %v", budgetID, err)
		return err
	}

	zap.L().Sugar().Infof("Cache invalidated for budgetID: %s", budgetID)
	return nil
}

func (b BudgetHistoryRepository) CreateInitialTX(ctx context.Context, tx *sqlx.Tx, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		zap.L().Sugar().Errorf("Failed to create initial transaction, budgetID: %s, error: %v", budgetID, err)
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create initial, budgetID: %s, error: %v", budgetID, err)
	}

	zap.L().Sugar().Infof("Initial transaction created with ID: %s for budgetID: %s", id, budgetID)
	return id, nil
}

func (b BudgetHistoryRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, budgetID, transactionID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance, transaction_id) VALUES ($1, $2, $3) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount, transactionID).Scan(&id); err != nil {
		zap.L().Sugar().Errorf("Failed to create transaction, budgetID: %s, transactionID: %s, error: %v", budgetID, transactionID, err)
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, budgetID: %s, transactionID: %s, error: %v", budgetID, transactionID, err)
	}

	zap.L().Sugar().Infof("Transaction created with ID: %s for budgetID: %s, transactionID: %s", id, budgetID, transactionID)
	return id, nil
}

func (b BudgetHistoryRepository) GetLastByBudgetID(ctx context.Context, budgetID string) (*domain.BudgetHistory, error) {
	if budgetID == "" {
		return nil, fmt.Errorf("budgetID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)
	fetch := func() (*domain.BudgetHistory, error) {
		var history domain.BudgetHistory
		query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
		if err := b.postgres.GetContext(ctx, &history, query, budgetID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch last budget history from DB, budgetID: %s, error: %v", budgetID, err)
			return nil, err
		}
		return &history, nil
	}

	result, err := db.WithCache(ctx, b.redis, cacheKey, TTL_GetLastHistoryByBudgetIDCache, fetch)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b BudgetHistoryRepository) Create(ctx context.Context, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := b.postgres.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		zap.L().Sugar().Errorf("Failed to create budget history, budgetID: %s, error: %v", budgetID, err)
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, budgetID: %s, error: %v", budgetID, err)
	}

	zap.L().Sugar().Infof("Budget history entry created with ID: %s for budgetID: %s", id, budgetID)
	return id, nil
}

func (b BudgetHistoryRepository) List(ctx context.Context, budgetID string) ([]*domain.BudgetHistory, error) {
	if budgetID == "" {
		return nil, fmt.Errorf("budgetID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyListHistory, budgetID)
	fetch := func() ([]*domain.BudgetHistory, error) {
		var histories []*domain.BudgetHistory
		query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at ASC", BudgetHistoryTable)
		if err := b.postgres.SelectContext(ctx, &histories, query, budgetID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch budget history list from DB, budgetID: %s, error: %v", budgetID, err)
			return nil, err
		}
		return histories, nil
	}

	result, err := db.WithCache(ctx, b.redis, cacheKey, TTL_ListBudgetHistoryCache, fetch)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b BudgetHistoryRepository) ListFromDate(ctx context.Context, budgetID string, fromDate time.Time, inclusive bool) ([]*domain.BudgetHistory, error) {
	if budgetID == "" {
		return nil, fmt.Errorf("budgetID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate.Unix(), inclusive)

	operator := ">"
	if inclusive {
		operator = ">="
	}

	fetch := func() ([]*domain.BudgetHistory, error) {
		var histories []*domain.BudgetHistory
		query := fmt.Sprintf(
			"SELECT * FROM %s WHERE budget_id = $1 AND created_at %s $2 ORDER BY created_at ASC",
			BudgetHistoryTable, operator,
		)
		if err := b.postgres.SelectContext(ctx, &histories, query, budgetID, fromDate); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch budget history from DB, budgetID: %s, fromDate: %v, error: %v", budgetID, fromDate, err)
			return nil, err
		}
		return histories, nil
	}

	result, err := db.WithCache(ctx, b.redis, cacheKey, TTL_ListBudgetHistoryFromDateCache, fetch)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b BudgetHistoryRepository) UpdateBalanceTX(ctx context.Context, tx *sqlx.Tx, transactionID string, amount float64) error {
	query := fmt.Sprintf("UPDATE %s SET balance = $1 WHERE transaction_id = $2", BudgetHistoryTable)

	if _, err := tx.ExecContext(ctx, query, amount, transactionID); err != nil {
		zap.L().Sugar().Errorf("Failed to update balance for transactionID: %s, error: %v", transactionID, err)
		return err
	}

	zap.L().Sugar().Infof("Balance updated for transactionID: %s", transactionID)
	return nil
}

func (b BudgetHistoryRepository) GetCurrentBalance(ctx context.Context, budgetID string) (float64, error) {
	if budgetID == "" {
		return 0, fmt.Errorf("budgetID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyBalance, budgetID)
	fetch := func() (float64, error) {
		var balance float64
		query := fmt.Sprintf("SELECT balance FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
		if err := b.postgres.GetContext(ctx, &balance, query, budgetID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch current balance from DB, budgetID: %s, error: %v", budgetID, err)
			return 0, err
		}
		return balance, nil
	}

	result, err := db.WithCache(ctx, b.redis, cacheKey, TTL_GetCurrentBalanceCache, fetch)
	if err != nil {
		return 0, err
	}

	return result, nil
}
