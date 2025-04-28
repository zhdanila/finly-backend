package repository

import (
	"context"
	"finly-backend/internal/domain"
	"finly-backend/internal/repository/auth"
	"finly-backend/internal/repository/budget"
	"finly-backend/internal/repository/budget_history"
	"finly-backend/internal/repository/category"
	"finly-backend/internal/repository/transaction"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"time"
)

type Repository struct {
	Auth
	Budget
	Category
	Transaction
	BudgetHistory
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

type Auth interface {
	Register(ctx context.Context, email, passwordHash, firstName, lastName string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)

	AddTokenToBlacklist(ctx context.Context, token string, ttlSeconds float64) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	RemoveToken(ctx context.Context, token string) error
}

type Budget interface {
	GetDB() *sqlx.DB
	CreateTX(ctx context.Context, tx *sqlx.Tx, userID, currency string) (string, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Budget, error)
}

type Category interface {
	Create(ctx context.Context, userID, name string) (string, error)
	GetByID(ctx context.Context, categoryID, userID string) (*domain.Category, error)
	List(ctx context.Context, userID string) ([]*domain.Category, error)
	ListCustom(ctx context.Context, userID string) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID, userID string) error
}

type Transaction interface {
	CreateTX(ctx context.Context, tx *sqlx.Tx, userID, budgetID, categoryID, transactionType, note string, amount float64) (string, error)
	GetDB() *sqlx.DB
	List(ctx context.Context, userID string) ([]*domain.Transaction, error)
	UpdateTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string, categoryID, transactionType, note string, amount float64) error
	DeleteTX(ctx context.Context, tx *sqlx.Tx, transactionID, userID string) error
	GetByID(ctx context.Context, transactionID, userID string) (*domain.Transaction, error)
}

type BudgetHistory interface {
	Create(ctx context.Context, budgetID string, amount float64) (string, error)
	CreateTX(ctx context.Context, tx *sqlx.Tx, budgetID, transactionID string, amount float64) (string, error)
	CreateInitialTX(ctx context.Context, tx *sqlx.Tx, budgetID string, amount float64) (string, error)
	GetLastByBudgetID(ctx context.Context, budgetID string) (*domain.BudgetHistory, error)
	List(ctx context.Context, budgetID string) ([]*domain.BudgetHistory, error)
	ListFromDate(ctx context.Context, budgetID string, fromDate time.Time, inclusive bool) ([]*domain.BudgetHistory, error)
	UpdateBalanceTX(ctx context.Context, tx *sqlx.Tx, transactionID string, amount float64) error
	GetCurrentBalance(ctx context.Context, budgetID string) (float64, error)
}
