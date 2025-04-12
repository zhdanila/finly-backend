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

func (a *AuthRepository) Register(ctx context.Context, email, passwordHash, firstName, lastName string) error {
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, firstName, lastName) VALUES ($1, $2, $3, $4)", UsersTable)

	_, err := a.db.Exec(query, email, passwordHash, firstName, lastName)
	if err != nil {
		return fmt.Errorf("could not insert buyer: %v", err)
	}

	return nil
}
