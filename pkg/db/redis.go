package db

import (
	"context"
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
