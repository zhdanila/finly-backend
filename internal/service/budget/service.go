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

	id, err := s.repo.Create(ctx, req.UserID, req.Currency)
	if err != nil {
		return nil, err
	}

	return &CreateBudgetResponse{
		ID: id,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, req *GetBudgetByIDRequest) (*GetBudgetByIDResponse, error) {
	var err error

	budget, err := s.repo.GetByID(ctx, req.BudgetID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetBudgetByIDResponse{
		Budget{
			ID:        budget.ID,
			UserID:    budget.UserID,
			Currency:  budget.Currency,
			CreatedAt: budget.CreatedAt,
			UpdatedAt: budget.UpdatedAt,
		},
	}, nil
}
