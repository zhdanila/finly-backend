package repository

import (
	"context"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/pkg/db"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

const (
	UsersTable = "users"

	TTL_GetUserCache = 30 * time.Minute

	cacheKeyUserByID    = "user:id:%s"
	cacheKeyUserByEmail = "user:email:%s"
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

func (a *AuthRepository) cacheKeys(userID, email string) []string {
	keys := []string{}
	if userID != "" {
		keys = append(keys, fmt.Sprintf(cacheKeyUserByID, userID))
	}
	if email != "" {
		keys = append(keys, fmt.Sprintf(cacheKeyUserByEmail, email))
	}
	return keys
}

func (a *AuthRepository) InvalidateCache(ctx context.Context, userID, email string) error {
	zap.L().Sugar().Infof("Invalidating cache for userID: %s, email: %s", userID, email)

	keys := a.cacheKeys(userID, email)
	if len(keys) == 0 {
		zap.L().Sugar().Info("No cache keys to invalidate")
		return nil
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
		zap.L().Sugar().Errorf("Failed to register user, email: %s, error: %v", email, err)
		return "", err
	}

	zap.L().Sugar().Infof("User registered, ID: %s, email: %s", userID, email)
	return userID, nil
}

func (a *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyUserByEmail, email)

	fetch := func() (*domain.User, error) {
		var user domain.User
		query := fmt.Sprintf("SELECT * FROM %s WHERE email = $1", UsersTable)
		if err := a.postgres.GetContext(ctx, &user, query, email); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch user from DB, email: %s, error: %v", email, err)
			return nil, err
		}
		zap.L().Sugar().Infof("Fetched user from DB by email: %s", email)
		return &user, nil
	}

	result, err := db.WithCache(ctx, a.redis, cacheKey, TTL_GetUserCache, fetch)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *AuthRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	cacheKey := fmt.Sprintf(cacheKeyUserByID, id)

	fetch := func() (*domain.User, error) {
		var user domain.User
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", UsersTable)
		if err := a.postgres.GetContext(ctx, &user, query, id); err != nil {
			zap.L().Sugar().Errorf("Failed to fetch user from DB, id: %s, error: %v", id, err)
			return nil, err
		}
		zap.L().Sugar().Infof("Fetched user from DB by ID: %s", id)
		return &user, nil
	}

	result, err := db.WithCache(ctx, a.redis, cacheKey, TTL_GetUserCache, fetch)
	if err != nil {
		return nil, err
	}

	return result, nil
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
		zap.L().Sugar().Infof("Token not blacklisted, token: %s", token)
		return false, nil
	} else if err != nil {
		zap.L().Sugar().Errorf("Redis error while checking blacklist, token: %s, error: %v", token, err)
		return false, err
	}
	zap.L().Sugar().Infof("Token blacklisted, token: %s", token)
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
