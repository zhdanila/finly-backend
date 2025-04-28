package category

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
	CategoryTable = "categories"

	TTL_ListCategoriesCache       = 1 * time.Hour
	TTL_ListCategoriesCustomCache = 1 * time.Hour
	TTL_GetCategoryCache          = 1 * time.Hour

	cacheKeyCategoryByIDAndUser    = "category:%s:user:%s"
	cacheKeyCategoriesByUser       = "categories:user:%s"
	cacheKeyCustomCategoriesByUser = "categories:custom:user:%s"
)

type CategoryRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewCategoryRepository(postgres *sqlx.DB, redis *redis.Client) *CategoryRepository {
	return &CategoryRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (c *CategoryRepository) cacheKeys(userID, categoryID string) []string {
	return []string{
		fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID),
		fmt.Sprintf(cacheKeyCategoriesByUser, userID),
		fmt.Sprintf(cacheKeyCustomCategoriesByUser, userID),
	}
}

func (c *CategoryRepository) InvalidateCache(ctx context.Context, userID, categoryID string) error {
	zap.L().Sugar().Infof("Invalidating category cache for userID: %s, categoryID: %s", userID, categoryID)

	keys := c.cacheKeys(userID, categoryID)
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate category cache for userID: %s, categoryID: %s, error: %v", userID, categoryID, err)
		return err
	}

	zap.L().Sugar().Infof("Category cache invalidated for userID: %s, categoryID: %s", userID, categoryID)
	return nil
}

func (c *CategoryRepository) Create(ctx context.Context, userID, name string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, name) VALUES ($1, $2) RETURNING id", CategoryTable)

	var id string
	if err := c.postgres.QueryRowContext(ctx, query, userID, name).Scan(&id); err != nil {
		zap.L().Sugar().Errorf("Failed to create category for userID: %s, name: %s, error: %v", userID, name, err)
		return "", err
	}

	if err := c.InvalidateCache(ctx, userID, id); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create for userID: %s, categoryID: %s, error: %v", userID, id, err)
	}

	zap.L().Sugar().Infof("Category created successfully for userID: %s, categoryID: %s", userID, id)
	return id, nil
}

func (c *CategoryRepository) GetByID(ctx context.Context, categoryID, userID string) (*domain.Category, error) {
	if categoryID == "" || userID == "" {
		return nil, fmt.Errorf("categoryID and userID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID)
	fetch := func() (*domain.Category, error) {
		var category domain.Category
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", CategoryTable)
		if err := c.postgres.GetContext(ctx, &category, query, categoryID, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch category from DB for categoryID: %s, userID: %s, error: %v", categoryID, userID, err)
			return nil, err
		}
		return &category, nil
	}

	return db.WithCache(ctx, c.redis, cacheKey, TTL_GetCategoryCache, fetch)
}

func (c *CategoryRepository) List(ctx context.Context, userID string) ([]*domain.Category, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyCategoriesByUser, userID)
	fetch := func() ([]*domain.Category, error) {
		var categories []*domain.Category
		query := fmt.Sprintf("SELECT * FROM %s WHERE (user_id = $1 AND is_user_category = true) OR is_user_category = false", CategoryTable)
		if err := c.postgres.SelectContext(ctx, &categories, query, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch categories from DB for userID: %s, error: %v", userID, err)
			return nil, err
		}
		return categories, nil
	}

	return db.WithCache(ctx, c.redis, cacheKey, TTL_ListCategoriesCache, fetch)
}

func (c *CategoryRepository) Delete(ctx context.Context, categoryID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2 AND is_user_category = true", CategoryTable)
	if _, err := c.postgres.ExecContext(ctx, query, categoryID, userID); err != nil {
		zap.L().Sugar().Errorf("Failed to delete category for categoryID: %s, userID: %s, error: %v", categoryID, userID, err)
		return err
	}

	if err := c.InvalidateCache(ctx, userID, categoryID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after delete for categoryID: %s, userID: %s, error: %v", categoryID, userID, err)
	}

	zap.L().Sugar().Infof("Deleted category for categoryID: %s, userID: %s", categoryID, userID)
	return nil
}

func (c *CategoryRepository) ListCustom(ctx context.Context, userID string) ([]*domain.Category, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyCustomCategoriesByUser, userID)
	fetch := func() ([]*domain.Category, error) {
		var categories []*domain.Category
		query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1 AND is_user_category = true", CategoryTable)
		if err := c.postgres.SelectContext(ctx, &categories, query, userID); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch custom categories from DB for userID: %s, error: %v", userID, err)
			return nil, err
		}
		return categories, nil
	}

	return db.WithCache(ctx, c.redis, cacheKey, TTL_ListCategoriesCustomCache, fetch)
}
