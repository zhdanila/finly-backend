package auth

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

func TestAuthRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Register", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			email := "test@example.com"
			passwordHash := "hashedpassword"
			firstName := "John"
			lastName := "Doe"
			userID := "123"

			query := fmt.Sprintf("INSERT INTO %s \\(email, password_hash, first_name, last_name\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id", UsersTable)
			mock.ExpectQuery(query).
				WithArgs(email, passwordHash, firstName, lastName).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

			id, err := repo.Register(ctx, email, passwordHash, firstName, lastName)
			assert.NoError(t, err)
			assert.Equal(t, userID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("DatabaseError", func(t *testing.T) {
			email := "test@example.com"
			passwordHash := "hashedpassword"
			firstName := "John"
			lastName := "Doe"

			query := fmt.Sprintf("INSERT INTO %s \\(email, password_hash, first_name, last_name\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id", UsersTable)
			mock.ExpectQuery(query).
				WithArgs(email, passwordHash, firstName, lastName).
				WillReturnError(errors.New("db error"))

			id, err := repo.Register(ctx, email, passwordHash, firstName, lastName)
			assert.Error(t, err)
			assert.Empty(t, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("CacheHit", func(t *testing.T) {
			email := "test@example.com"
			cacheKey := fmt.Sprintf(cacheKeyUserByEmail, email)
			user := &domain.User{ID: "123", Email: email, FirstName: "John"}

			data, err := json.Marshal(user)
			assert.NoError(t, err)

			err = redisClient.Set(ctx, cacheKey, data, TTL_GetUserCache).Err()
			assert.NoError(t, err)

			result, err := repo.GetUserByEmail(ctx, email)
			assert.NoError(t, err)
			assert.Equal(t, user, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("CacheMiss", func(t *testing.T) {
			email := "test2@example.com"
			user := &domain.User{ID: "123", Email: email, FirstName: "John"}

			query := fmt.Sprintf("SELECT \\* FROM %s WHERE email = \\$1", UsersTable)
			mock.ExpectQuery(query).
				WithArgs(email).
				WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name"}).
					AddRow(user.ID, user.Email, user.FirstName))

			result, err := repo.GetUserByEmail(ctx, email)
			assert.NoError(t, err)
			assert.Equal(t, user, result)
			assert.NoError(t, mock.ExpectationsWereMet())

			cached, err := redisClient.Get(ctx, fmt.Sprintf(cacheKeyUserByEmail, email)).Result()
			assert.NoError(t, err)
			assert.NotEmpty(t, cached)
		})

		t.Run("EmptyEmail", func(t *testing.T) {
			result, err := repo.GetUserByEmail(ctx, "")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			userID := "123"
			email := "test@example.com"
			cacheKeyID := fmt.Sprintf(cacheKeyUserByID, userID)
			cacheKeyEmail := fmt.Sprintf(cacheKeyUserByEmail, email)

			redisClient.Set(ctx, cacheKeyID, "data", 0)
			redisClient.Set(ctx, cacheKeyEmail, "data", 0)

			err := repo.InvalidateCache(ctx, userID, email)
			assert.NoError(t, err)

			existsID, _ := redisClient.Exists(ctx, cacheKeyID).Result()
			existsEmail, _ := redisClient.Exists(ctx, cacheKeyEmail).Result()
			assert.Equal(t, int64(0), existsID)
			assert.Equal(t, int64(0), existsEmail)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("NoKeys", func(t *testing.T) {
			err := repo.InvalidateCache(ctx, "", "")
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("AddTokenToBlacklist", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			token := "testtoken"
			ttl := 3600.0

			err := repo.AddTokenToBlacklist(ctx, token, ttl)
			assert.NoError(t, err)

			val, err := redisClient.Get(ctx, token).Result()
			assert.NoError(t, err)
			assert.Equal(t, "blacklisted", val)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("IsTokenBlacklisted", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("Blacklisted", func(t *testing.T) {
			token := "testtoken"
			redisClient.Set(ctx, token, "blacklisted", 0)

			isBlacklisted, err := repo.IsTokenBlacklisted(ctx, token)
			assert.NoError(t, err)
			assert.True(t, isBlacklisted)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("NotBlacklisted", func(t *testing.T) {
			token := "testtoken2"

			isBlacklisted, err := repo.IsTokenBlacklisted(ctx, token)
			assert.NoError(t, err)
			assert.False(t, isBlacklisted)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("RemoveToken", func(t *testing.T) {
		sqlxDB, mock, redisClient, mr, logger := testutil.SetupRepositoryTest(t)
		defer sqlxDB.Close()
		defer mr.Close()

		zap.ReplaceGlobals(logger)
		repo := NewAuthRepository(sqlxDB, redisClient)

		t.Run("Success", func(t *testing.T) {
			token := "testtoken"
			redisClient.Set(ctx, token, "blacklisted", 0)

			err := repo.RemoveToken(ctx, token)
			assert.NoError(t, err)

			exists, _ := redisClient.Exists(ctx, token).Result()
			assert.Equal(t, int64(0), exists)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
