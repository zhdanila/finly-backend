package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type TokenBlacklistRepository struct {
	redis *redis.Client
}

func NewTokenBlacklistRepository(redis *redis.Client) *TokenBlacklistRepository {
	return &TokenBlacklistRepository{
		redis: redis,
	}
}

func (t TokenBlacklistRepository) AddToken(ctx context.Context, token string, ttlSeconds float64) error {
	return t.redis.Set(ctx, token, "blacklisted", time.Duration(ttlSeconds)*time.Second).Err()
}

func (t TokenBlacklistRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	exists, err := t.redis.Exists(ctx, token).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
