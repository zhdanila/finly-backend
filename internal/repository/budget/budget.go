package budget

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
	BudgetTable = "budgets"

	TTL_GetBudgetByUserIDCache = 30 * time.Minute

	cacheKeyBudgetByUser = "budget:user:%s"
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

func (b *BudgetRepository) GetDB() *sqlx.DB {
	return b.postgres
}

func (b *BudgetRepository) cacheKeys(userID string) []string {
	return []string{
		fmt.Sprintf(cacheKeyBudgetByUser, userID),
	}
}

func (b *BudgetRepository) InvalidateCache(ctx context.Context, userID string) error {
	zap.L().Sugar().Infof("Invalidating cache for userID: %s", userID)

	keys := b.cacheKeys(userID)
	if err := b.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache for userID: %s, error: %v", userID, err)
		return err
	}

	zap.L().Sugar().Infof("Cache invalidated for userID: %s", userID)
	return nil
}

func (b *BudgetRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, userID, currency string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, currency) VALUES ($1, $2) RETURNING id", BudgetTable)

	var id string
	if err := tx.QueryRowContext(ctx, query, userID, currency).Scan(&id); err != nil {
		zap.L().Sugar().Errorf("Failed to create budget for userID: %s, currency: %s, error: %v", userID, currency, err)
		return "", err
	}

	if err := b.InvalidateCache(ctx, userID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create, userID: %s, error: %v", userID, err)
	} else {
		zap.L().Sugar().Infof("Cache invalidated after create for userID: %s", userID)
	}

	return id, nil
}

func (b *BudgetRepository) GetByUserID(ctx context.Context, userID string) (*domain.Budget, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyBudgetByUser, userID)

	fetch := func() (*domain.Budget, error) {
		var budget domain.Budget
		query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", BudgetTable)
		if err := b.postgres.GetContext(ctx, &budget, query, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch budget from DB, userID: %s, error: %v", userID, err)
			return nil, err
		}
		zap.L().Sugar().Infof("Fetched budget from DB for userID: %s, budgetID: %s", userID, budget.ID)
		return &budget, nil
	}

	result, err := db.WithCache(ctx, b.redis, cacheKey, TTL_GetBudgetByUserIDCache, fetch)
	if err != nil {
		return nil, err
	}

	zap.L().Sugar().Infof("Fetched budget with cache for userID: %s", userID)
	return result, nil
}
