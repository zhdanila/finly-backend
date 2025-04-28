package testutil

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func SetupRepositoryTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *redis.Client, *miniredis.Miniredis, *zap.Logger) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	logger, _ := zap.NewDevelopment()

	return sqlxDB, mock, redisClient, mr, logger
}
