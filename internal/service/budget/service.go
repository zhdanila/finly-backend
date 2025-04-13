package budget

import (
	"context"
	"finly-backend/internal/repository"
)

type Service struct {
	repo repository.Budget
}

func NewService(repo repository.Budget) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateBudgetRequest) (*CreateBudgetResponse, error) {
	var err error

	if err = s.repo.Create(ctx, req.UserID, req.Name, req.Currency); err != nil {
		return nil, err
	}

	return &CreateBudgetResponse{}, nil
}

func (s *Service) GetByID(ctx context.Context, req *GetBudgetByIDRequest) (*GetBudgetByIDResponse, error) {
	var err error

	budget, err := s.repo.GetByID(ctx, req.BudgetID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetBudgetByIDResponse{
		Budget{
			Name:     budget.Name,
			Amount:   budget.Amount.Int64,
			Currency: budget.Currency,
		},
	}, nil
}

func (s *Service) List(ctx context.Context, req *ListBudgetsByIDRequest) (*ListBudgetsByIDResponse, error) {
	var err error

	budgets, err := s.repo.List(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	budgetsResponse := make([]Budget, len(budgets))
	for i, budget := range budgets {
		budgetsResponse[i] = Budget{
			Name:     budget.Name,
			Amount:   budget.Amount.Int64,
			Currency: budget.Currency,
		}
	}

	return &ListBudgetsByIDResponse{
		Budgets: budgetsResponse,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *DeleteBudgetRequest) (*DeleteBudgetResponse, error) {
	var err error

	if err = s.repo.Delete(ctx, req.BudgetID, req.UserID); err != nil {
		return nil, err
	}

	return &DeleteBudgetResponse{}, nil
}
