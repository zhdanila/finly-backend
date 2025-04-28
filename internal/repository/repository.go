package repository

import (
	"finly-backend/internal/repository/auth"
	"finly-backend/internal/repository/budget"
	"finly-backend/internal/repository/budget_history"
	"finly-backend/internal/repository/category"
	"finly-backend/internal/repository/transaction"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	auth.Auth
	budget.Budget
	category.Category
	transaction.Transaction
	budget_history.BudgetHistory
}

func NewRepository(postgres *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Auth:          auth.NewAuthRepository(postgres, redis),
		Budget:        budget.NewBudgetRepository(postgres, redis),
		Category:      category.NewCategoryRepository(postgres, redis),
		Transaction:   transaction.NewTransactionRepository(postgres, redis),
		BudgetHistory: budget_history.NewBudgetHistoryRepository(postgres, redis),
	}
}
