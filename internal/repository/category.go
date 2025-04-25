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
	CategoryTable = "categories"

	TTL_ListCategoriesCache       = 1 * time.Hour
	TTL_ListCategoriesCustomCache = 1 * time.Hour
	TTL_GetCategoryCache          = 1 * time.Hour
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

func (c CategoryRepository) InvalidateCache(ctx context.Context, userID, categoryID string) error {
	keys := []string{
		fmt.Sprintf("category:%s:user:%s", categoryID, userID),
		fmt.Sprintf("categories:user:%s", userID),
		fmt.Sprintf("categories:custom:user:%s", userID),
	}
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache for userID: %s, categoryID: %s, error: %v", userID, categoryID, err)
		return err
	}
	zap.L().Sugar().Infof("Cache invalidated for userID: %s, categoryID: %s", userID, categoryID)
	return nil
}

func (c CategoryRepository) Create(ctx context.Context, userID, name string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, name) VALUES ($1, $2) RETURNING id", CategoryTable)

	var id string
	if err := c.postgres.QueryRowContext(ctx, query, userID, name).Scan(&id); err != nil {
		return "", err
	}

	if err := c.InvalidateCache(ctx, userID, id); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after create: %v", err)
	}

	return id, nil
}

func (c CategoryRepository) GetByID(ctx context.Context, categoryID, userID string) (*domain.Category, error) {
	cacheKey := fmt.Sprintf("category:%s:user:%s", categoryID, userID)

	var category domain.Category
	cachedCategory, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedCategory), &category); err == nil {
			return &category, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error: %v\n", err)
	} else {
		zap.L().Sugar().Infof("Cache miss for category, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", CategoryTable)
	if err = c.postgres.GetContext(ctx, &category, query, categoryID, userID); err != nil {
		return nil, err
	}

	serializedCategories, err := json.Marshal(category)
	if err == nil {
		if err = c.redis.Set(ctx, cacheKey, serializedCategories, TTL_GetCategoryCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache categories: %v", err)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal category for caching, categoryID: %s, userID: %s, error: %v", categoryID, userID, err)
	}

	return &category, nil
}

func (c CategoryRepository) List(ctx context.Context, userID string) ([]*domain.Category, error) {
	cacheKey := fmt.Sprintf("categories:user:%s", userID)

	var categories []*domain.Category
	cachedCategories, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedCategories), &categories); err == nil {
			return categories, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error: %v\n", err)
	} else {
		zap.L().Sugar().Infof("Cache miss for categories, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE (user_id = $1 AND is_user_category = true) OR is_user_category = false", CategoryTable)
	if err = c.postgres.SelectContext(ctx, &categories, query, userID); err != nil {
		return nil, err
	}

	serializedCategories, err := json.Marshal(categories)
	if err == nil {
		if err = c.redis.Set(ctx, cacheKey, serializedCategories, TTL_ListCategoriesCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache categories: %v", err)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal categories for caching, userID: %s, error: %v", userID, err)
	}

	return categories, nil
}

func (c CategoryRepository) Delete(ctx context.Context, categoryID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2 AND is_user_category = true", CategoryTable)

	if _, err := c.postgres.ExecContext(ctx, query, categoryID, userID); err != nil {
		return err
	}

	if err := c.InvalidateCache(ctx, userID, categoryID); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache after delete: %v", err)
	}

	return nil
}

func (c CategoryRepository) ListCustom(ctx context.Context, userID string) ([]*domain.Category, error) {
	cacheKey := fmt.Sprintf("categories:custom:user:%s", userID)

	var categories []*domain.Category
	cachedCategories, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedCategories), &categories); err == nil {
			return categories, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error: %v\n", err)
	} else {
		zap.L().Sugar().Infof("Cache miss for custom categories, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1 AND is_user_category = true", CategoryTable)
	if err = c.postgres.SelectContext(ctx, &categories, query, userID); err != nil {
		return nil, err
	}

	serializedCategories, err := json.Marshal(categories)
	if err == nil {
		if err = c.redis.Set(ctx, cacheKey, serializedCategories, TTL_ListCategoriesCustomCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache categories: %v", err)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal categories for caching, userID: %s, error: %v", userID, err)
	}

	return categories, nil
}
