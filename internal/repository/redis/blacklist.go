package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type TokenRepository struct {
	redis *redis.Client
}

func NewTokenRepository(redis *redis.Client) *TokenRepository {
	return &TokenRepository{
		redis: redis,
	}
}

func (t TokenRepository) AddTokenToBlacklist(ctx context.Context, token string, ttlSeconds float64) error {
	return t.redis.Set(ctx, token, "blacklisted", time.Duration(ttlSeconds)*time.Second).Err()
}

func (t TokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := t.redis.Get(ctx, token).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "blacklisted", nil
}

func (t TokenRepository) RemoveToken(ctx context.Context, token string) error {
	return t.redis.Del(ctx, token).Err()
}
