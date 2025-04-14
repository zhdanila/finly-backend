package service

import (
	"finly-backend/internal/repository"
	"finly-backend/internal/service/auth"
	"finly-backend/internal/service/budget"
	"finly-backend/internal/service/category"
	"finly-backend/internal/service/transaction"
)

type Service struct {
	Auth        *auth.Service
	Budget      *budget.Service
	Category    *category.Service
	Transaction *transaction.Service
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Auth:        auth.NewService(repos.Auth, repos.Budget),
		Budget:      budget.NewService(repos.Budget),
		Category:    category.NewService(repos.Category),
		Transaction: transaction.NewService(repos.Transaction, repos.BudgetHistory),
	}
}
