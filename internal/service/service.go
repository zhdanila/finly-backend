package service

import (
	"finly-backend/internal/repository"
	"finly-backend/internal/service/auth"
	"finly-backend/internal/service/budget"
	"finly-backend/internal/service/category"
	"finly-backend/internal/service/transaction"
	transactionExec "finly-backend/pkg/transaction"
)

type Service struct {
	Auth        auth.Auth
	Budget      budget.Budget
	Category    category.Category
	Transaction transaction.Transaction
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Auth:        auth.NewService(repos.Auth, repos.Budget),
		Budget:      budget.NewService(repos.Budget, repos.BudgetHistory, transactionExec.NewTransactionExecutor()),
		Category:    category.NewService(repos.Category),
		Transaction: transaction.NewService(repos.Transaction, repos.BudgetHistory, transactionExec.NewTransactionExecutor()),
	}
}
