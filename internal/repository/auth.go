package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (a *AuthRepository) Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id", UsersTable)

	var userID string
	err := a.db.QueryRowContext(ctx, query, email, passwordHash, firstName, lastName).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("could not insert user: %v", err)
	}

	return userID, nil
}
