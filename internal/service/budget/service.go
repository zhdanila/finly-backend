package budget

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/repository/budget"
	"finly-backend/internal/repository/budget_history"
	"finly-backend/pkg/db"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Budget interface {
	Create(ctx context.Context, req *CreateBudgetRequest) (*CreateBudgetResponse, error)
	GetByUserID(ctx context.Context, req *GetBudgetByIDRequest) (*GetBudgetByIDResponse, error)
	GetBudgetHistory(ctx context.Context, req *GetBudgetHistoryRequest) (*GetBudgetHistoryResponse, error)
	GetCurrentBalance(ctx context.Context, req *GetCurrentBalanceRequest) (*GetCurrentBalanceResponse, error)
}

type Service struct {
	budgetRepo        budget.Budget
	budgetHistoryRepo budget_history.BudgetHistory
}

func NewService(budgetRepo budget.Budget, budgetHistoryRepo budget_history.BudgetHistory) *Service {
	return &Service{
		budgetRepo:        budgetRepo,
		budgetHistoryRepo: budgetHistoryRepo,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateBudgetRequest) (*CreateBudgetResponse, error) {
	var budgetID string
	err := db.WithTransaction(ctx, s.budgetRepo.GetDB(), func(tx *sqlx.Tx) error {
		var err error

		budgetID, err = s.budgetRepo.CreateTX(ctx, tx, req.UserID, req.Currency)
		if err != nil {
			zap.L().Sugar().Errorf("Create: failed to create budget: %v", err)
			return err
		}

		if req.Amount != 0 {
			if _, err = s.budgetHistoryRepo.CreateInitialTX(ctx, tx, budgetID, req.Amount); err != nil {
				zap.L().Sugar().Errorf("Create: failed to create initial history for budgetID=%s: %v", budgetID, err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		zap.L().Sugar().Errorf("Create: transaction failed for userID=%s: %v", req.UserID, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Create: completed for userID=%s", req.UserID)
	return &CreateBudgetResponse{ID: budgetID}, nil
}

func (s *Service) GetByUserID(ctx context.Context, req *GetBudgetByIDRequest) (*GetBudgetByIDResponse, error) {
	budget, err := s.budgetRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Sugar().Infof("GetByUserID: no budget found for userID=%s", req.UserID)
			return &GetBudgetByIDResponse{}, nil
		}
		zap.L().Sugar().Errorf("GetByUserID: failed for userID=%s: %v", req.UserID, err)
		return nil, err
	}

	return &GetBudgetByIDResponse{
		&BudgetObject{
			ID:        budget.ID,
			UserID:    budget.UserID,
			Currency:  budget.Currency,
			CreatedAt: budget.CreatedAt,
			UpdatedAt: budget.UpdatedAt,
		},
	}, nil
}

func (s *Service) GetBudgetHistory(ctx context.Context, req *GetBudgetHistoryRequest) (*GetBudgetHistoryResponse, error) {
	budgets, err := s.budgetHistoryRepo.List(ctx, req.BudgetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Sugar().Infof("GetBudgetHistory: no history for budgetID=%s", req.BudgetID)
			return &GetBudgetHistoryResponse{}, nil
		}
		zap.L().Sugar().Errorf("GetBudgetHistory: failed for budgetID=%s: %v", req.BudgetID, err)
		return nil, err
	}

	var history []*BudgetHistory
	for _, b := range budgets {
		history = append(history, &BudgetHistory{
			ID:        b.ID,
			BudgetID:  b.BudgetID,
			Balance:   b.Balance,
			CreatedAt: b.CreatedAt,
		})
	}

	return &GetBudgetHistoryResponse{
		BudgetHistory: history,
	}, nil
}

func (s *Service) GetCurrentBalance(ctx context.Context, req *GetCurrentBalanceRequest) (*GetCurrentBalanceResponse, error) {
	balance, err := s.budgetHistoryRepo.GetCurrentBalance(ctx, req.BudgetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Sugar().Infof("GetCurrentBalance: no balance for budgetID=%s", req.BudgetID)
			return &GetCurrentBalanceResponse{}, nil
		}
		zap.L().Sugar().Errorf("GetCurrentBalance: failed for budgetID=%s: %v", req.BudgetID, err)
		return nil, err
	}

	return &GetCurrentBalanceResponse{
		Balance: balance,
	}, nil
}
