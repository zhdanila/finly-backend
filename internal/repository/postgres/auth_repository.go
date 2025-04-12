package postgres

import (
	"context"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
)

const UsersTable = "users"

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
		return "", err
	}

	return userID, nil
}

func (a *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT password_hash FROM %s WHERE email = $1", UsersTable)

	var user domain.User
	if err := a.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", UsersTable)

	var user domain.User
	if err := a.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}
