package budget

import (
	"context"
	"finly-backend/internal/repository"
	"strings"
)

type Service struct {
	budgetRepo        repository.Budget
	budgetHistoryRepo repository.BudgetHistory
}

func NewService(budgetRepo repository.Budget, budgetHistoryRepo repository.BudgetHistory) *Service {
	return &Service{
		budgetRepo:        budgetRepo,
		budgetHistoryRepo: budgetHistoryRepo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateBudgetRequest) (*CreateBudgetResponse, error) {
	var err error

	budgetID, err := s.budgetRepo.Create(ctx, req.UserID, req.Currency)
	if err != nil {
		return nil, err
	}

	if req.Amount != 0 {
		if _, err = s.budgetHistoryRepo.Create(ctx, budgetID, req.Amount); err != nil {
			return nil, err
		}
	}

	return &CreateBudgetResponse{
		ID: budgetID,
	}, nil
}

func (s *Service) GetByUserID(ctx context.Context, req *GetBudgetByIDRequest) (*GetBudgetByIDResponse, error) {
	var err error

	budget, err := s.budgetRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &GetBudgetByIDResponse{}, nil
		}
		return nil, err
	}

	return &GetBudgetByIDResponse{
		&Budget{
			ID:        budget.ID,
			UserID:    budget.UserID,
			Currency:  budget.Currency,
			CreatedAt: budget.CreatedAt,
			UpdatedAt: budget.UpdatedAt,
		},
	}, nil
}

func (s *Service) GetBudgetHistory(ctx context.Context, req *GetBudgetHistoryRequest) (*GetBudgetHistoryResponse, error) {
	var err error

	budgets, err := s.budgetHistoryRepo.List(ctx, req.BudgetID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &GetBudgetHistoryResponse{}, nil
		}
		return nil, err
	}

	var budgetHistory []*BudgetHistory
	for _, budget := range budgets {
		budgetHistory = append(budgetHistory, &BudgetHistory{
			ID:        budget.ID,
			BudgetID:  budget.BudgetID,
			Balance:   budget.Balance,
			CreatedAt: budget.CreatedAt,
		})
	}

	return &GetBudgetHistoryResponse{
		BudgetHistory: budgetHistory,
	}, nil
}
