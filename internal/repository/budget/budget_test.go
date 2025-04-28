package budget

import (
	"context"
	"encoding/json"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/pkg/testutil"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestBudgetRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("GetDB", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			db := repo.GetDB()
			assert.Equal(t, sqlxDB, db)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("cacheKeys", func(t *testing.T) {
		sqlxDB, _, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			expected := []string{fmt.Sprintf(cacheKeyBudgetByUser, userID)}
			keys := repo.cacheKeys(userID)
			assert.Equal(t, expected, keys)
		})
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			cacheKey := fmt.Sprintf(cacheKeyBudgetByUser, userID)

			redisClient.Set(ctx, cacheKey, "data", 0)

			err := repo.InvalidateCache(ctx, userID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("RedisError", func(t *testing.T) {
			userID := "123"

			mr.Close()

			err := repo.InvalidateCache(ctx, userID)
			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("CreateTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			currency := "USD"
			budgetID := "456"
			cacheKey := fmt.Sprintf(cacheKeyBudgetByUser, userID)

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(user_id, currency\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetTable)
			mock.ExpectQuery(query).
				WithArgs(userID, currency).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(budgetID))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, userID, currency)
			assert.NoError(t, err)
			assert.Equal(t, budgetID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			userID := "123"
			currency := "USD"

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(user_id, currency\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetTable)
			mock.ExpectQuery(query).
				WithArgs(userID, currency).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, userID, currency)
			assert.Error(t, err)
			assert.Empty(t, id)

			err = tx.Rollback()
			assert.NoError(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())

		})

		t.Run("CacheInvalidateError", func(t *testing.T) {
			userID := "123"
			currency := "USD"
			budgetID := "456"

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(user_id, currency\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetTable)
			mock.ExpectQuery(query).
				WithArgs(userID, currency).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(budgetID))

			mr.Close()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, userID, currency)
			assert.NoError(t, err)
			assert.Equal(t, budgetID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetByUserID", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := "123"
			cacheKey := fmt.Sprintf(cacheKeyBudgetByUser, userID)
			budget := &domain.Budget{ID: "456", UserID: userID, Currency: "USD"}

			data, err := json.Marshal(budget)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetBudgetByUserIDCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetByUserID(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, budget, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "1234"
			budget := &domain.Budget{ID: "456", UserID: userID, Currency: "USD"}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE user_id = \\$1", BudgetTable)
			mock.ExpectQuery(query).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "currency"}).
					AddRow(budget.ID, budget.UserID, budget.Currency))

			result, err := repo.GetByUserID(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, budget, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyBudgetByUser, userID)).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyUserID", func(t *testing.T) {
			result, err := repo.GetByUserID(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			userID := "12345"

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE user_id = \\$1", BudgetTable)
			mock.ExpectQuery(query).
				WithArgs(userID).
				WillReturnError(errors.New("db error"))

			result, err := repo.GetByUserID(ctx, userID)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
