package category

import (
	"context"
	"database/sql"
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

func TestCategoryRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("cacheKeys", func(t *testing.T) {
		sqlxDB, _, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			categoryID := "456"
			expected := []string{
				fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID),
				fmt.Sprintf(cacheKeyCategoriesByUser, userID),
				fmt.Sprintf(cacheKeyCustomCategoriesByUser, userID),
			}
			keys := repo.cacheKeys(userID, categoryID)
			assert.Equal(t, expected, keys)
		})
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			categoryID := "456"
			cacheKey := fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID)
			redisClient.Set(ctx, cacheKey, "data", 0)

			err := repo.InvalidateCache(ctx, userID, categoryID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("RedisError", func(t *testing.T) {
			userID := "123"
			categoryID := "456"
			mr.Close()

			err := repo.InvalidateCache(ctx, userID, categoryID)
			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("Create", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			name := "Test CategoryObject"
			categoryID := "456"
			cacheKey := fmt.Sprintf(cacheKeyCategoriesByUser, userID)

			query := fmt.Sprintf("INSERT INTO %s \\(user_id, name\\) VALUES \\(\\$1, \\$2\\) RETURNING id", CategoryTable)
			mock.ExpectQuery(query).
				WithArgs(userID, name).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(categoryID))

			id, err := repo.Create(ctx, userID, name)
			assert.NoError(t, err)
			assert.Equal(t, categoryID, id)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			userID := "123"
			name := "Test CategoryObject"

			query := fmt.Sprintf("INSERT INTO %s \\(user_id, name\\) VALUES \\(\\$1, \\$2\\) RETURNING id", CategoryTable)
			mock.ExpectQuery(query).
				WithArgs(userID, name).
				WillReturnError(errors.New("db error"))

			id, err := repo.Create(ctx, userID, name)
			assert.Error(t, err)
			assert.Empty(t, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := sql.NullString{String: "123"}
			categoryID := "456"
			cacheKey := fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID.String)
			category := &domain.Category{ID: categoryID, UserID: userID, Name: "Test CategoryObject"}

			data, err := json.Marshal(category)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetCategoryCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetByID(ctx, categoryID, userID.String)
			assert.NoError(t, err)
			assert.Equal(t, category, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "1234"
			categoryID := "4567"
			category := &domain.Category{
				ID: categoryID,
				UserID: sql.NullString{
					String: userID,
					Valid:  true,
				},
				Name: "Test CategoryObject",
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE id = \\$1 AND user_id = \\$2", CategoryTable)
			mock.ExpectQuery(query).
				WithArgs(categoryID, userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "is_user_category"}).
					AddRow(category.ID, category.UserID.String, category.Name, false))

			result, err := repo.GetByID(ctx, categoryID, userID)
			assert.NoError(t, err)
			assert.Equal(t, category, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID)).Result()
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

	t.Run("List", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := sql.NullString{String: "12345"}
			cacheKey := fmt.Sprintf(cacheKeyCategoriesByUser, userID.String)
			categories := []*domain.Category{
				{ID: "456", UserID: userID, Name: "Test CategoryObject"},
			}

			data, err := json.Marshal(categories)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_ListCategoriesCache).Err()
			assert.NoError(t, err)

			result, err := repo.List(ctx, userID.String)
			assert.NoError(t, err)
			assert.Equal(t, categories, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "1234567"
			categories := []*domain.Category{
				{
					ID: "456",
					UserID: sql.NullString{
						String: userID,
						Valid:  true,
					},
					Name:           "Test CategoryObject",
					IsUserCategory: true,
				},
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE \\(user_id = \\$1 AND is_user_category = true\\) OR is_user_category = false", CategoryTable)
			mock.ExpectQuery(query).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "is_user_category"}).
					AddRow(categories[0].ID, categories[0].UserID.String, categories[0].Name, categories[0].IsUserCategory))

			result, err := repo.List(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, categories, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyCategoriesByUser, userID)).Result()
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

	t.Run("Delete", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			categoryID := "456"
			cacheKey := fmt.Sprintf(cacheKeyCategoryByIDAndUser, categoryID, userID)

			query := fmt.Sprintf("DELETE FROM %s WHERE id = \\$1 AND user_id = \\$2 AND is_user_category = true", CategoryTable)
			mock.ExpectExec(query).
				WithArgs(categoryID, userID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.Delete(ctx, categoryID, userID)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, cacheKey).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			userID := "123"
			categoryID := "456"

			query := fmt.Sprintf("DELETE FROM %s WHERE id = \\$1 AND user_id = \\$2 AND is_user_category = true", CategoryTable)
			mock.ExpectExec(query).
				WithArgs(categoryID, userID).
				WillReturnError(errors.New("db error"))

			err := repo.Delete(ctx, categoryID, userID)
			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("ListCustom", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewCategoryRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			userID := sql.NullString{String: "12345"}
			cacheKey := fmt.Sprintf(cacheKeyCustomCategoriesByUser, userID.String)
			categories := []*domain.Category{
				{ID: "456", UserID: userID, Name: "Test CategoryObject", IsUserCategory: true},
			}

			data, err := json.Marshal(categories)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_ListCategoriesCustomCache).Err()
			assert.NoError(t, err)

			result, err := repo.ListCustom(ctx, userID.String)
			assert.NoError(t, err)
			assert.Equal(t, categories, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			userID := "1234567"
			categories := []*domain.Category{
				{
					ID: "456",
					UserID: sql.NullString{
						String: userID,
						Valid:  true,
					},
					Name:           "Test CategoryObject",
					IsUserCategory: true,
				},
			}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE user_id = \\$1 AND is_user_category = true", CategoryTable)
			mock.ExpectQuery(query).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "is_user_category"}).
					AddRow(categories[0].ID, categories[0].UserID.String, categories[0].Name, categories[0].IsUserCategory))

			result, err := repo.ListCustom(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, categories, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyCustomCategoriesByUser, userID)).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyUserID", func(t *testing.T) {
			result, err := repo.ListCustom(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
