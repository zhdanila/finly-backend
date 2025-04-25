package repository

import (
	"context"
	"encoding/json"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

const (
	BudgetTable                = "budgets"
	TTL_GetBudgetByUserIDCache = 30 * time.Minute
)

type BudgetRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewBudgetRepository(postgres *sqlx.DB, redis *redis.Client) *BudgetRepository {
	return &BudgetRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (b BudgetRepository) InvalidateCache(ctx context.Context, userID string) error {
	cacheKey := fmt.Sprintf("budget:user:%s", userID)
	if err := b.redis.Del(ctx, cacheKey).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache for userID: %s, error: %v", userID, err)
		return err
	}
	zap.L().Sugar().Infof("Cache invalidated for userID: %s", userID)
	return nil
}

func (b BudgetRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, userID, currency string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, currency) VALUES ($1, $2) RETURNING id", BudgetTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, userID, currency).Scan(&id); err != nil {
		return "", err
	}

	if err := b.InvalidateCache(ctx, userID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, userID: %s, error: %v", userID, err)
	}

	return id, nil
}

func (b BudgetRepository) GetByUserID(ctx context.Context, userID string) (*domain.Budget, error) {
	cacheKey := fmt.Sprintf("budget:user:%s", userID)

	var budget domain.Budget
	cachedBudget, err := b.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedBudget), &budget); err == nil {
			zap.L().Sugar().Infof("Cache hit for budget, key: %s", cacheKey)
			return &budget, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached budget, key: %s, error: %v", cacheKey, err)
	} else if err != redis.Nil {
		zap.L().Sugar().Errorf("Redis error for key: %s, userID: %s, error: %v", cacheKey, userID, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for budget, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", BudgetTable)
	if err = b.postgres.GetContext(ctx, &budget, query, userID); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch budget from DB, userID: %s, error: %v", userID, err)
		return nil, err
	}

	serializedBudget, err := json.Marshal(budget)
	if err == nil {
		if err = b.redis.Set(ctx, cacheKey, serializedBudget, TTL_GetBudgetByUserIDCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache budget, key: %s, userID: %s, error: %v", cacheKey, userID, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached budget, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal budget for caching, userID: %s, error: %v", userID, err)
	}

	return &budget, nil
}

func (b BudgetRepository) GetDB() *sqlx.DB {
	return b.postgres
}
