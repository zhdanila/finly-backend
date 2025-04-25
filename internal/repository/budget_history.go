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
	BudgetHistoryTable = "budgets_history"

	TTL_GetLastHistoryByBudgetIDCache  = 5 * time.Minute
	TTL_ListBudgetHistoryCache         = 15 * time.Minute
	TTL_ListBudgetHistoryFromDateCache = 15 * time.Minute
	TTL_GetCurrentBalanceCache         = 30 * time.Second
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

func (b BudgetHistoryRepository) InvalidateCache(ctx context.Context, budgetID string) error {
	keys := []string{
		fmt.Sprintf("budget:history:last:%s", budgetID),
		fmt.Sprintf("budget:history:list:%s", budgetID),
		fmt.Sprintf("budget:balance:%s", budgetID),
		fmt.Sprintf("budget:history:from:%s:*", budgetID),
	}
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
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create initial, budgetID: %s, error: %v", budgetID, err)
	}

	return id, nil
}

func (b BudgetHistoryRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, budgetID, transactionID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance, transaction_id) VALUES ($1, $2, $3) RETURNING id", BudgetHistoryTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, budgetID, amount, transactionID).Scan(&id); err != nil {
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, budgetID: %s, transactionID: %s, error: %v", budgetID, transactionID, err)
	}

	return id, nil
}

func (b BudgetHistoryRepository) GetLastByBudgetID(ctx context.Context, budgetID string) (*domain.BudgetHistory, error) {
	cacheKey := fmt.Sprintf("budget:history:last:%s", budgetID)

	var history domain.BudgetHistory
	cachedHistory, err := b.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedHistory), &history); err == nil {
			zap.L().Sugar().Infof("Cache hit for last budget history, key: %s", cacheKey)
			return &history, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached budget history, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for last budget history, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
	if err = b.postgres.GetContext(ctx, &history, query, budgetID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch last budget history from DB, budgetID: %s, error: %v", budgetID, err)
		return nil, err
	}

	serializedHistory, err := json.Marshal(history)
	if err == nil {
		if err = b.redis.Set(ctx, cacheKey, serializedHistory, TTL_GetLastHistoryByBudgetIDCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache last budget history, key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached last budget history, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal budget history for caching, budgetID: %s, error: %v", budgetID, err)
	}

	return &history, nil
}

func (b BudgetHistoryRepository) Create(ctx context.Context, budgetID string, amount float64) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (budget_id, balance) VALUES ($1, $2) RETURNING id", BudgetHistoryTable)

	var id string
	if err := b.postgres.QueryRowContext(ctx, query, budgetID, amount).Scan(&id); err != nil {
		return "", err
	}

	if err := b.InvalidateCache(ctx, budgetID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, budgetID: %s, error: %v", budgetID, err)
	}

	return id, nil
}

func (b BudgetHistoryRepository) List(ctx context.Context, budgetID string) ([]*domain.BudgetHistory, error) {
	cacheKey := fmt.Sprintf("budget:history:list:%s", budgetID)

	var histories []*domain.BudgetHistory
	cachedHistories, err := b.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedHistories), &histories); err == nil {
			zap.L().Sugar().Infof("Cache hit for budget history list, key: %s", cacheKey)
			return histories, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached budget history list, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for budget history list, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE budget_id = $1 ORDER BY created_at ASC", BudgetHistoryTable)
	if err = b.postgres.SelectContext(ctx, &histories, query, budgetID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch budget history list from DB, budgetID: %s, error: %v", budgetID, err)
		return nil, err
	}

	serializedHistories, err := json.Marshal(histories)
	if err == nil {
		if err = b.redis.Set(ctx, cacheKey, serializedHistories, TTL_ListBudgetHistoryCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache budget history list, key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached budget history list, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal budget history list for caching, budgetID: %s, error: %v", budgetID, err)
	}

	return histories, nil
}

func (b BudgetHistoryRepository) ListFromDate(ctx context.Context, budgetID string, fromDate time.Time, inclusive bool) ([]*domain.BudgetHistory, error) {
	cacheKey := fmt.Sprintf("budget:history:from:%s:date:%d:inclusive:%t", budgetID, fromDate.Unix(), inclusive)

	var histories []*domain.BudgetHistory
	cachedHistories, err := b.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedHistories), &histories); err == nil {
			zap.L().Sugar().Infof("Cache hit for budget history from date, key: %s", cacheKey)
			return histories, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached budget history from date, key: %s, error: %v", cacheKey, err)
	} else if err != redis.Nil {
		zap.L().Sugar().Errorf("Redis error for key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for budget history from date, key: %s", cacheKey)
	}

	operator := ">"
	if inclusive {
		operator = ">="
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE budget_id = $1 AND created_at %s $2 ORDER BY created_at ASC",
		BudgetHistoryTable, operator,
	)
	if err = b.postgres.SelectContext(ctx, &histories, query, budgetID, fromDate); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch budget history from date from DB, budgetID: %s, fromDate: %v, error: %v", budgetID, fromDate, err)
		return nil, err
	}

	serializedHistories, err := json.Marshal(histories)
	if err == nil {
		if err = b.redis.Set(ctx, cacheKey, serializedHistories, TTL_ListBudgetHistoryFromDateCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache budget history from date, key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached budget history from date, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal budget history from date for caching, budgetID: %s, error: %v", budgetID, err)
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
	cacheKey := fmt.Sprintf("budget:balance:%s", budgetID)

	var balance float64
	cachedBalance, err := b.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedBalance), &balance); err == nil {
			zap.L().Sugar().Infof("Cache hit for current balance, key: %s", cacheKey)
			return balance, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached current balance, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for current balance, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT balance FROM %s WHERE budget_id = $1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
	if err = b.postgres.GetContext(ctx, &balance, query, budgetID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch current balance from DB, budgetID: %s, error: %v", budgetID, err)
		return 0, err
	}

	serializedBalance, err := json.Marshal(balance)
	if err == nil {
		if err = b.redis.Set(ctx, cacheKey, serializedBalance, TTL_GetCurrentBalanceCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache current balance, key: %s, budgetID: %s, error: %v", cacheKey, budgetID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached current balance, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal current balance for caching, budgetID: %s, error: %v", budgetID, err)
	}

	return balance, nil
}
