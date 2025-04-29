package budget_history

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
	"time"
)

func TestBudgetHistoryRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("cacheKeys", func(t *testing.T) {
		sqlxDB, _, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("WithoutFromDate", func(t *testing.T) {
			budgetID := "123"
			expected := []string{
				fmt.Sprintf(cacheKeyLastHistory, budgetID),
				fmt.Sprintf(cacheKeyListHistory, budgetID),
				fmt.Sprintf(cacheKeyBalance, budgetID),
			}
			keys := repo.cacheKeys(budgetID)
			assert.Equal(t, expected, keys)
		})

		t.Run("WithFromDate", func(t *testing.T) {
			budgetID := "123"
			fromDate := time.Now()
			expected := []string{
				fmt.Sprintf(cacheKeyLastHistory, budgetID),
				fmt.Sprintf(cacheKeyListHistory, budgetID),
				fmt.Sprintf(cacheKeyBalance, budgetID),
				fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate.Unix(), true),
				fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate.Unix(), false),
			}
			keys := repo.cacheKeys(budgetID, fromDate)
			assert.Equal(t, expected, keys)
		})
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			budgetID := "123"
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)
			redisClient.Set(ctx, cacheKey, "data", 0)

			err := repo.InvalidateCache(ctx, budgetID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("RedisError", func(t *testing.T) {
			budgetID := "123"
			mr.Close()

			err := repo.InvalidateCache(ctx, budgetID)
			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("CreateInitialTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			budgetID := "123"
			amount := 100.0
			historyID := "456"
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateInitialTX(ctx, tx, budgetID, amount)
			assert.NoError(t, err)
			assert.Equal(t, historyID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			budgetID := "123"
			amount := 100.0

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateInitialTX(ctx, tx, budgetID, amount)
			assert.Error(t, err)
			assert.Empty(t, id)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("CreateTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			budgetID := "123"
			transactionID := "789"
			amount := 100.0
			historyID := "456"
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance, transaction_id\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount, transactionID).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, budgetID, transactionID, amount)
			assert.NoError(t, err)
			assert.Equal(t, historyID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			budgetID := "123"
			transactionID := "789"
			amount := 100.0

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance, transaction_id\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount, transactionID).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, budgetID, transactionID, amount)
			assert.Error(t, err)
			assert.Empty(t, id)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetLastByBudgetID", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			budgetID := "123"
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)
			history := &domain.BudgetHistory{ID: "456", BudgetID: budgetID, Balance: 100.0}

			data, err := json.Marshal(history)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetLastHistoryByBudgetIDCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetLastByBudgetID(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, history, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			budgetID := "1234"
			history := &domain.BudgetHistory{ID: "456", BudgetID: budgetID, Balance: 100.0}
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE budget_id = \\$1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "budget_id", "balance"}).
					AddRow(history.ID, history.BudgetID, history.Balance))

			result, err := repo.GetLastByBudgetID(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, history, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, cacheKey).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyBudgetID", func(t *testing.T) {
			result, err := repo.GetLastByBudgetID(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("Create", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			budgetID := "123"
			amount := 100.0
			historyID := "456"
			cacheKey := fmt.Sprintf(cacheKeyLastHistory, budgetID)

			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))

			id, err := repo.Create(ctx, budgetID, amount)
			assert.NoError(t, err)
			assert.Equal(t, historyID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			budgetID := "123"
			amount := 100.0

			query := fmt.Sprintf("INSERT INTO %s \\(budget_id, balance\\) VALUES \\(\\$1, \\$2\\) RETURNING id", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, amount).
				WillReturnError(errors.New("db error"))

			id, err := repo.Create(ctx, budgetID, amount)
			assert.Error(t, err)
			assert.Empty(t, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("List", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			budgetID := "123"
			cacheKey := fmt.Sprintf(cacheKeyListHistory, budgetID)
			histories := []*domain.BudgetHistory{
				{ID: "456", BudgetID: budgetID, Balance: 100.0},
			}

			data, err := json.Marshal(histories)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_ListBudgetHistoryCache).Err()
			assert.NoError(t, err)

			result, err := repo.List(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, histories, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			budgetID := "12345"
			cacheKey := fmt.Sprintf(cacheKeyListHistory, budgetID)
			histories := []*domain.BudgetHistory{
				{ID: "456", BudgetID: budgetID, Balance: 100.0},
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE budget_id = \\$1 ORDER BY created_at ASC", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "budget_id", "balance"}).
					AddRow(histories[0].ID, histories[0].BudgetID, histories[0].Balance))

			result, err := repo.List(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, histories, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, cacheKey).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyBudgetID", func(t *testing.T) {
			result, err := repo.List(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("ListFromDate", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			budgetID := "123"
			fromDate := time.Now()
			inclusive := true
			cacheKey := fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate.Unix(), inclusive)
			histories := []*domain.BudgetHistory{
				{ID: "456", BudgetID: budgetID, Balance: 100.0},
			}

			data, err := json.Marshal(histories)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_ListBudgetHistoryFromDateCache).Err()
			assert.NoError(t, err)

			result, err := repo.ListFromDate(ctx, budgetID, fromDate, inclusive)
			assert.NoError(t, err)
			assert.Equal(t, histories, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			budgetID := "12345"
			fromDate := time.Now()
			inclusive := true
			histories := []*domain.BudgetHistory{
				{ID: "456", BudgetID: budgetID, Balance: 100.0},
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE budget_id = \\$1 AND created_at \\>= \\$2 ORDER BY created_at ASC", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID, fromDate).
				WillReturnRows(sqlmock.NewRows([]string{"id", "budget_id", "balance"}).
					AddRow(histories[0].ID, histories[0].BudgetID, histories[0].Balance))

			result, err := repo.ListFromDate(ctx, budgetID, fromDate, inclusive)
			assert.NoError(t, err)
			assert.Equal(t, histories, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyHistoryFrom, budgetID, fromDate.Unix(), inclusive)).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyBudgetID", func(t *testing.T) {
			result, err := repo.ListFromDate(ctx, "", time.Now(), true)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("UpdateBalanceTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			transactionID := "789"
			amount := 100.0

			mock.ExpectBegin()
			query := fmt.Sprintf("UPDATE %s SET balance = \\$1 WHERE transaction_id = \\$2", BudgetHistoryTable)
			mock.ExpectExec(query).
				WithArgs(amount, transactionID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.UpdateBalanceTX(ctx, tx, transactionID, amount)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			transactionID := "789"
			amount := 100.0

			mock.ExpectBegin()
			query := fmt.Sprintf("UPDATE %s SET balance = \\$1 WHERE transaction_id = \\$2", BudgetHistoryTable)
			mock.ExpectExec(query).
				WithArgs(amount, transactionID).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.UpdateBalanceTX(ctx, tx, transactionID, amount)
			assert.Error(t, err)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetCurrentBalance", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewBudgetHistoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			budgetID := "123"
			cacheKey := fmt.Sprintf(cacheKeyBalance, budgetID)
			balance := 100.0

			data, err := json.Marshal(balance)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetCurrentBalanceCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetCurrentBalance(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, balance, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			budgetID := "12456"
			balance := 100.0

			query := fmt.Sprintf("SELECT balance FROM %s WHERE budget_id = \\$1 ORDER BY created_at DESC LIMIT 1", BudgetHistoryTable)
			mock.ExpectQuery(query).
				WithArgs(budgetID).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(balance))

			result, err := repo.GetCurrentBalance(ctx, budgetID)
			assert.NoError(t, err)
			assert.Equal(t, balance, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyBalance, budgetID)).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyBudgetID", func(t *testing.T) {
			result, err := repo.GetCurrentBalance(ctx, "")
			assert.Error(t, err)
			assert.Equal(t, 0.0, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
