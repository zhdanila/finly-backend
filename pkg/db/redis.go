package db

import (
	"context"
	"encoding/json"
	"errors"
	"finly-backend/internal/config"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisDB(ctx context.Context, cnf *config.Config) (*redis.Client, error) {
	redisConfig := RedisConfig{
		Addr:     fmt.Sprintf("%s:%s", cnf.RedisHost, cnf.RedisPort),
		Password: cnf.RedisPassword,
		DB:       cnf.RedisDB,
	}

	var client *redis.Client
	var err error

	for i := 0; i < 10; i++ {
		client = redis.NewClient(&redis.Options{
			Addr:     redisConfig.Addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		_, err = client.Ping(ctx).Result()
		if err == nil {
			zap.L().Info("Redis connected")
			return client, nil
		}

		zap.L().Error("Redis not ready, retrying in 5 seconds...", zap.Error(err))
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("unable to connect to Redis after 10 attempts: %w", err)
}

func GetFromCache[T any](ctx context.Context, redisClient *redis.Client, key string, dest *T) (bool, error) {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // cache miss
		}
		return false, err // redis error
	}
	if err = json.Unmarshal([]byte(val), dest); err != nil {
		return false, err // unmarshalling error
	}
	return true, nil // cache hit
}

func SetToCache(ctx context.Context, redis *redis.Client, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redis.Set(ctx, key, data, ttl).Err()
}

func WithCache[T any](ctx context.Context, redisClient *redis.Client, cacheKey string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	var result T
	cacheHit, err := GetFromCache(ctx, redisClient, cacheKey, &result)
	if err != nil {
		zap.L().Error("Cache error", zap.String("cacheKey", cacheKey), zap.Error(err))
	} else if cacheHit {
		zap.L().Info("Cache hit", zap.String("cacheKey", cacheKey))
		return result, nil
	}

	data, err := fetch()
	if err != nil {
		return result, err
	}

	if err = SetToCache(ctx, redisClient, cacheKey, data, ttl); err != nil {
		zap.L().Error("Failed to set cache", zap.String("cacheKey", cacheKey), zap.Error(err))
	}

	zap.L().Info("Cache miss, data fetched and cached", zap.String("cacheKey", cacheKey))
	return data, nil
}
