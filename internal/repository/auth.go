package repository

import (
	"context"
	"errors"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"time"
)

const UsersTable = "users"

type AuthRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewAuthRepository(postgres *sqlx.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (a *AuthRepository) Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id", UsersTable)

	var userID string
	err := a.postgres.QueryRowContext(ctx, query, email, passwordHash, firstName, lastName).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (a *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE email = $1", UsersTable)

	var user domain.User
	if err := a.postgres.GetContext(ctx, &user, query, email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", UsersTable)

	var user domain.User
	if err := a.postgres.GetContext(ctx, &user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthRepository) AddTokenToBlacklist(ctx context.Context, token string, ttlSeconds float64) error {
	return a.redis.Set(ctx, token, "blacklisted", time.Duration(ttlSeconds)*time.Second).Err()
}

func (a *AuthRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := a.redis.Get(ctx, token).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "blacklisted", nil
}

func (a *AuthRepository) RemoveToken(ctx context.Context, token string) error {
	return a.redis.Del(ctx, token).Err()
}
