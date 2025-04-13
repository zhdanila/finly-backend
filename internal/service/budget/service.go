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
