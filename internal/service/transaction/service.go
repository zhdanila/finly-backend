package transaction

import (
	"context"
	"finly-backend/internal/repository"
	"strings"
)

type Service struct {
	repo repository.Transaction
}

func NewService(repo repository.Transaction) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	var err error

	id, err := s.repo.Create(ctx, req.UserID, req.BudgetID, req.CategoryID, req.Type.String(), req.Note, req.Amount)
	if err != nil {
		if strings.Contains(err.Error(), "not enough budget") {
			return nil, errs.NotEnoughBudget
		}
		return nil, err
	}

	return &CreateTransactionResponse{
		ID: id,
	}, nil
}
