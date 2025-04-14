package repository

import (
	"context"
	"finly-backend/internal/domain"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const CategoryTable = "categories"

type CategoryRepository struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

func NewCategoryRepository(postgres *sqlx.DB, redis *redis.Client) *CategoryRepository {
	return &CategoryRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (c CategoryRepository) Create(ctx context.Context, userID, name, description string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, name, description) VALUES ($1, $2, $3) RETURNING id", CategoryTable)

	var id string
	if err := c.postgres.QueryRowContext(ctx, query, userID, name, description).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (c CategoryRepository) GetByID(ctx context.Context, categoryID, userID string) (*domain.Category, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND user_id = $2", CategoryTable)

	var category domain.Category
	if err := c.postgres.GetContext(ctx, &category, query, categoryID, userID); err != nil {
		return nil, err
	}

	return &category, nil
}

func (c CategoryRepository) List(ctx context.Context, userID string) ([]*domain.Category, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = $1", CategoryTable)

	var categories []*domain.Category
	if err := c.postgres.SelectContext(ctx, &categories, query, userID); err != nil {
		return nil, err
	}

	return categories, nil
}

func (c CategoryRepository) Delete(ctx context.Context, categoryID, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2", CategoryTable)

	if _, err := c.postgres.ExecContext(ctx, query, categoryID, userID); err != nil {
		return err
	}

	return nil
}
