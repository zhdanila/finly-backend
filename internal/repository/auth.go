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
	UsersTable = "users"

	TTL_GetUserCache = 30 * time.Minute
)

type AuthRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewAuthRepository(postgres *sqlx.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (a *AuthRepository) InvalidateCache(ctx context.Context, userID, email string) error {
	keys := []string{
		fmt.Sprintf("user:id:%s", userID),
		fmt.Sprintf("user:email:%s", email),
	}
	if err := a.redis.Del(ctx, keys...).Err(); err != nil {
		zap.L().Sugar().Warnf("Failed to invalidate cache for userID: %s, email: %s, error: %v", userID, email, err)
		return err
	}
	zap.L().Sugar().Infof("Cache invalidated for userID: %s, email: %s", userID, email)
	return nil
}

func (a *AuthRepository) Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id", UsersTable)

	var userID string
	err := a.postgres.QueryRowContext(ctx, query, email, passwordHash, firstName, lastName).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (a *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)

	var user domain.User
	cachedUser, err := a.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedUser), &user); err == nil {
			zap.L().Sugar().Infof("Cache hit for user by email, key: %s", cacheKey)
			return &user, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached user, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, email: %s, error: %v", cacheKey, email, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for user by email, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE email = $1", UsersTable)
	if err = a.postgres.GetContext(ctx, &user, query, email); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch user from DB, email: %s, error: %v", email, err)
		return nil, err
	}

	serializedUser, err := json.Marshal(user)
	if err == nil {
		if err = a.redis.Set(ctx, cacheKey, serializedUser, TTL_GetUserCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache user, key: %s, email: %s, error: %v", cacheKey, email, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached user, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal user for caching, email: %s, error: %v", email, err)
	}

	return &user, nil
}

func (a *AuthRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:id:%s", id)

	var user domain.User
	cachedUser, err := a.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedUser), &user); err == nil {
			zap.L().Sugar().Infof("Cache hit for user by ID, key: %s", cacheKey)
			return &user, nil
		}
		zap.L().Sugar().Warnf("Failed to unmarshal cached user, key: %s, error: %v", cacheKey, err)
	} else if !errors.Is(err, redis.Nil) {
		zap.L().Sugar().Errorf("Redis error for key: %s, id: %s, error: %v", cacheKey, id, err)
	} else {
		zap.L().Sugar().Infof("Cache miss for user by ID, key: %s", cacheKey)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", UsersTable)
	if err = a.postgres.GetContext(ctx, &user, query, id); err != nil {
		zap.L().Sugar().Errorf("Failed to fetch user from DB, id: %s, error: %v", id, err)
		return nil, err
	}

	serializedUser, err := json.Marshal(user)
	if err == nil {
		if err = a.redis.Set(ctx, cacheKey, serializedUser, TTL_GetUserCache).Err(); err != nil {
			zap.L().Sugar().Errorf("Failed to cache user, key: %s, id: %s, error: %v", cacheKey, id, err)
		} else {
			zap.L().Sugar().Infof("Successfully cached user, key: %s", cacheKey)
		}
	} else {
		zap.L().Sugar().Warnf("Failed to marshal user for caching, id: %s, error: %v", id, err)
	}

	return &user, nil
}

func (a *AuthRepository) AddTokenToBlacklist(ctx context.Context, token string, ttlSeconds float64) error {
	if err := a.redis.Set(ctx, token, "blacklisted", time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		zap.L().Sugar().Errorf("Failed to add token to blacklist, token: %s, error: %v", token, err)
		return err
	}
	zap.L().Sugar().Infof("Token added to blacklist, token: %s", token)
	return nil
}

func (a *AuthRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := a.redis.Get(ctx, token).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "blacklisted", nil
}

func (a *AuthRepository) RemoveToken(ctx context.Context, token string) error {
	if err := a.redis.Del(ctx, token).Err(); err != nil {
		zap.L().Sugar().Errorf("Failed to remove token from blacklist, token: %s, error: %v", token, err)
		return err
	}
	zap.L().Sugar().Infof("Token removed from blacklist, token: %s", token)
	return nil
}
