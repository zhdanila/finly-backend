package transaction

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

func TestTransactionRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("cacheKeys", func(t *testing.T) {
		sqlxDB, _, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			transactionID := "456"
			expected := []string{
				fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID),
				fmt.Sprintf(cacheKeyTransactionsByUser, userID),
			}
			keys := repo.cacheKeys(userID, transactionID)
			assert.Equal(t, expected, keys)
		})
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			transactionID := "456"
			cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)
			redisClient.Set(ctx, cacheKey, "data", 0)

			err := repo.InvalidateCache(ctx, userID, transactionID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("RedisError", func(t *testing.T) {
			userID := "123"
			transactionID := "456"
			mr.Close()

			err := repo.InvalidateCache(ctx, userID, transactionID)
			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetDB", func(t *testing.T) {
		sqlxDB, _, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			db := repo.GetDB()
			assert.Equal(t, sqlxDB, db)
		})
	})

	t.Run("CreateTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			budgetID := "789"
			categoryID := "101"
			transactionType := "expense"
			note := "Test transaction"
			amount := 50.0
			transactionID := "456"
			cacheKey := fmt.Sprintf(cacheKeyTransactionsByUser, userID)

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(user_id, budget_id, category_id, amount, transaction_type, note\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\) RETURNING id", TransactionTable)
			mock.ExpectQuery(query).
				WithArgs(userID, budgetID, categoryID, amount, transactionType, note).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(transactionID))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, userID, budgetID, categoryID, transactionType, note, amount)
			assert.NoError(t, err)
			assert.Equal(t, transactionID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			userID := "123"
			budgetID := "789"
			categoryID := "101"
			transactionType := "expense"
			note := "Test transaction"
			amount := 50.0

			mock.ExpectBegin()
			query := fmt.Sprintf("INSERT INTO %s \\(user_id, budget_id, category_id, amount, transaction_type, note\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\) RETURNING id", TransactionTable)
			mock.ExpectQuery(query).
				WithArgs(userID, budgetID, categoryID, amount, transactionType, note).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			id, err := repo.CreateTX(ctx, tx, userID, budgetID, categoryID, transactionType, note, amount)
			assert.Error(t, err)
			assert.Empty(t, id)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("List", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := "123"
			cacheKey := fmt.Sprintf(cacheKeyTransactionsByUser, userID)
			transactions := []*domain.Transaction{
				{
					ID:              "456",
					UserID:          userID,
					BudgetID:        "789",
					CategoryID:      "101",
					Amount:          50.0,
					TransactionType: "expense",
					Note:            "Test transaction",
				},
			}

			data, err := json.Marshal(transactions)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_ListTransactionsCache).Err()
			assert.NoError(t, err)

			result, err := repo.List(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, transactions, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "12345"
			cacheKey := fmt.Sprintf(cacheKeyTransactionsByUser, userID)
			transactions := []*domain.Transaction{
				{
					ID:              "456",
					UserID:          userID,
					BudgetID:        "789",
					CategoryID:      "101",
					Amount:          50.0,
					TransactionType: "expense",
					Note:            "Test transaction",
				},
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE user_id = \\$1 ORDER BY created_at DESC", TransactionTable)
			mock.ExpectQuery(query).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "budget_id", "category_id", "amount", "transaction_type", "note"}).
					AddRow(transactions[0].ID, transactions[0].UserID, transactions[0].BudgetID, transactions[0].CategoryID, transactions[0].Amount, transactions[0].TransactionType, transactions[0].Note))

			result, err := repo.List(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, transactions, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, cacheKey).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyUserID", func(t *testing.T) {
			result, err := repo.List(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("UpdateTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			transactionID := "456"
			userID := "123"
			categoryID := "101"
			transactionType := "expense"
			note := "Updated transaction"
			amount := 75.0
			cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)

			mock.ExpectBegin()
			query := fmt.Sprintf("UPDATE %s SET category_id = \\$1, transaction_type = \\$2, note = \\$3, amount = \\$4 WHERE id = \\$5 AND user_id = \\$6", TransactionTable)
			mock.ExpectExec(query).
				WithArgs(categoryID, transactionType, note, amount, transactionID, userID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.UpdateTX(ctx, tx, transactionID, userID, categoryID, transactionType, note, amount)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			transactionID := "456"
			userID := "123"
			categoryID := "101"
			transactionType := "expense"
			note := "Updated transaction"
			amount := 75.0

			mock.ExpectBegin()
			query := fmt.Sprintf("UPDATE %s SET category_id = \\$1, transaction_type = \\$2, note = \\$3, amount = \\$4 WHERE id = \\$5 AND user_id = \\$6", TransactionTable)
			mock.ExpectExec(query).
				WithArgs(categoryID, transactionType, note, amount, transactionID, userID).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.UpdateTX(ctx, tx, transactionID, userID, categoryID, transactionType, note, amount)
			assert.Error(t, err)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("DeleteTX", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			transactionID := "456"
			userID := "123"
			cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)

			mock.ExpectBegin()
			query := fmt.Sprintf("DELETE FROM %s WHERE id = \\$1 AND user_id = \\$2", TransactionTable)
			mock.ExpectExec(query).
				WithArgs(transactionID, userID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.DeleteTX(ctx, tx, transactionID, userID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			transactionID := "456"
			userID := "123"

			mock.ExpectBegin()
			query := fmt.Sprintf("DELETE FROM %s WHERE id = \\$1 AND user_id = \\$2", TransactionTable)
			mock.ExpectExec(query).
				WithArgs(transactionID, userID).
				WillReturnError(errors.New("db error"))
			mock.ExpectRollback()

			tx, err := sqlxDB.Beginx()
			assert.NoError(t, err)

			err = repo.DeleteTX(ctx, tx, transactionID, userID)
			assert.Error(t, err)

			err = tx.Rollback()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewTransactionRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := "123"
			transactionID := "456"
			cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)
			transaction := &domain.Transaction{
				ID:              transactionID,
				UserID:          userID,
				BudgetID:        "789",
				CategoryID:      "101",
				Amount:          50.0,
				TransactionType: "expense",
				Note:            "Test transaction",
			}

			data, err := json.Marshal(transaction)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetByIDTransactionCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetByID(ctx, transactionID, userID)
			assert.NoError(t, err)
			assert.Equal(t, transaction, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "12345"
			transactionID := "4567"
			cacheKey := fmt.Sprintf(cacheKeyTransactionByIDAndUser, transactionID, userID)
			transaction := &domain.Transaction{
				ID:              transactionID,
				UserID:          userID,
				BudgetID:        "789",
				CategoryID:      "101",
				Amount:          50.0,
				TransactionType: "expense",
				Note:            "Test transaction",
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE id = \\$1 AND user_id = \\$2", TransactionTable)
			mock.ExpectQuery(query).
				WithArgs(transactionID, userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "budget_id", "category_id", "amount", "transaction_type", "note"}).
					AddRow(transaction.ID, transaction.UserID, transaction.BudgetID, transaction.CategoryID, transaction.Amount, transaction.TransactionType, transaction.Note))

			result, err := repo.GetByID(ctx, transactionID, userID)
			assert.NoError(t, err)
			assert.Equal(t, transaction, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, cacheKey).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyInputs", func(t *testing.T) {
			result, err := repo.GetByID(ctx, "", "123")
			assert.Error(t, err)
			assert.Nil(t, result)

			result, err = repo.GetByID(ctx, "456", "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
