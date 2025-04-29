package transaction

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"finly-backend/internal/repository/budget_history"
	"finly-backend/internal/repository/transaction"
	transactionExec "finly-backend/pkg/transaction"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type Transaction interface {
	Create(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error)
	List(ctx context.Context, req *ListTransactionRequest) (*ListTransactionResponse, error)
	Update(ctx context.Context, req *UpdateTransactionRequest) (*UpdateTransactionResponse, error)
	Delete(ctx context.Context, req *DeleteTransactionRequest) (*DeleteTransactionResponse, error)
}

type Service struct {
	transactionRepo   transaction.Transaction
	budgetHistoryRepo budget_history.BudgetHistory

	transactionExecutor transactionExec.TransactionExecutor
}

func NewService(transactionRepo transaction.Transaction, budgetHistoryRepo budget_history.BudgetHistory, transactionExecutor transactionExec.TransactionExecutor) *Service {
	return &Service{
		transactionRepo:     transactionRepo,
		budgetHistoryRepo:   budgetHistoryRepo,
		transactionExecutor: transactionExecutor,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	var (
		transactionID string
		err           error
	)

	if err = s.transactionExecutor.WithTransaction(ctx, s.transactionRepo.GetDB(), func(tx *sqlx.Tx) error {
		transactionID, err = s.transactionRepo.CreateTX(ctx, tx, req.UserID, req.BudgetID, req.CategoryID, req.Type.String(), req.Note, req.Amount)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to create transaction for userID=%s, budgetID=%s: %v", req.UserID, req.BudgetID, err)
			return errs.DatabaseError
		}

		lastBudgetHistory, err := s.budgetHistoryRepo.GetLastByBudgetID(ctx, req.BudgetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				lastBudgetHistory = nil
			} else {
				zap.L().Sugar().Errorf("Failed to get last budget history for budgetID=%s: %v", req.BudgetID, err)
				return errs.DatabaseError
			}
		}

		newAmount, err := calculateNewAmount(lastBudgetHistory, req.Amount, req.Type.String())
		if err != nil {
			zap.L().Sugar().Errorf("Failed to calculate new amount for budgetID=%s: %v", req.BudgetID, err)
			return errs.InvalidInput
		}

		_, err = s.budgetHistoryRepo.CreateTX(ctx, tx, req.BudgetID, transactionID, newAmount)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to create budget history for transactionID=%s, budgetID=%s: %v", transactionID, req.BudgetID, err)
			return errs.DatabaseError
		}

		return nil
	}); err != nil {
		zap.L().Sugar().Errorf("TransactionObject creation failed for userID=%s, budgetID=%s: %v", req.UserID, req.BudgetID, err)
		return nil, err
	}

	return &CreateTransactionResponse{ID: transactionID}, nil
}

func (s *Service) List(ctx context.Context, req *ListTransactionRequest) (*ListTransactionResponse, error) {
	transactions, err := s.transactionRepo.List(ctx, req.UserID)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to list transactions for userID=%s: %v", req.UserID, err)
		return nil, errs.DatabaseError
	}

	transactionList := make([]TransactionObject, 0, len(transactions))
	for _, t := range transactions {
		transactionList = append(transactionList, TransactionObject{
			ID:         t.ID,
			UserID:     t.UserID,
			BudgetID:   t.BudgetID,
			CategoryID: t.CategoryID,
			Type:       e_transaction_type.Enum(t.TransactionType),
			Note:       t.Note,
			Amount:     t.Amount,
			CreatedAt:  t.CreatedAt,
		})
	}

	return &ListTransactionResponse{Transactions: transactionList}, nil
}

func (s *Service) Update(ctx context.Context, req *UpdateTransactionRequest) (*UpdateTransactionResponse, error) {
	if err := s.transactionExecutor.WithTransaction(ctx, s.transactionRepo.GetDB(), func(tx *sqlx.Tx) error {
		if err := s.transactionRepo.UpdateTX(ctx, tx, req.TransactionID, req.UserID, req.CategoryID, req.Type, req.Note, req.Amount); err != nil {
			zap.L().Sugar().Errorf("Failed to update transactionID=%s for userID=%s: %v", req.TransactionID, req.UserID, err)
			return errs.DatabaseError
		}

		transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID, req.UserID)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to get transaction by ID for transactionID=%s, userID=%s: %v", req.TransactionID, req.UserID, err)
			return errs.DatabaseError
		}

		if transaction.TransactionType != req.Type || transaction.Amount != req.Amount {
			difference, err := calculateDeltaChange(transaction.TransactionType, transaction.Amount, req.Type, req.Amount)
			if err != nil {
				zap.L().Sugar().Errorf("Failed to calculate delta change for transactionID=%s: %v", req.TransactionID, err)
				return errs.InvalidTransactionType
			}

			if err = s.updateBudgetHistory(ctx, tx, transaction.BudgetID, transaction.CreatedAt, difference, true); err != nil {
				zap.L().Sugar().Errorf("Failed to update budget history for transactionID=%s: %v", req.TransactionID, err)
				return err
			}
		}

		return nil
	}); err != nil {
		zap.L().Sugar().Errorf("TransactionObject update failed for transactionID=%s, userID=%s: %v", req.TransactionID, req.UserID, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Successfully updated transactionID=%s for userID=%s", req.TransactionID, req.UserID)
	return &UpdateTransactionResponse{}, nil
}

func (s *Service) Delete(ctx context.Context, req *DeleteTransactionRequest) (*DeleteTransactionResponse, error) {
	if err := s.transactionExecutor.WithTransaction(ctx, s.transactionRepo.GetDB(), func(tx *sqlx.Tx) error {
		transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID, req.UserID)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to get transaction by ID for transactionID=%s, userID=%s: %v", req.TransactionID, req.UserID, err)
			return errs.DatabaseError
		}

		difference, err := invertDelta(transaction.TransactionType, transaction.Amount, false)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to invert delta for transactionID=%s: %v", req.TransactionID, err)
			return errs.InvalidTransactionType
		}

		if err = s.updateBudgetHistory(ctx, tx, transaction.BudgetID, transaction.CreatedAt, difference, false); err != nil {
			zap.L().Sugar().Errorf("Failed to update budget history for transactionID=%s: %v", req.TransactionID, err)
			return err
		}

		if err = s.transactionRepo.DeleteTX(ctx, tx, req.TransactionID, req.UserID); err != nil {
			zap.L().Sugar().Errorf("Failed to delete transactionID=%s for userID=%s: %v", req.TransactionID, req.UserID, err)
			return errs.DatabaseError
		}

		return nil
	}); err != nil {
		zap.L().Sugar().Errorf("TransactionObject deletion failed for transactionID=%s, userID=%s: %v", req.TransactionID, req.UserID, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Successfully deleted transactionID=%s for userID=%s", req.TransactionID, req.UserID)
	return &DeleteTransactionResponse{}, nil
}

func (s *Service) updateBudgetHistory(ctx context.Context, tx *sqlx.Tx, budgetID string, fromDate time.Time, difference float64, inclusiveDate bool) error {
	budgetHistory, err := s.budgetHistoryRepo.ListFromDate(ctx, budgetID, fromDate, inclusiveDate)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to list budget history for budgetID=%s: %v", budgetID, err)
		return errs.DatabaseError
	}

	for _, history := range budgetHistory {
		newAmount := history.Balance + difference
		if newAmount < 0 {
			zap.L().Sugar().Errorf("Insufficient balance for budgetID=%s, transactionID=%s", budgetID, history.TransactionID.String)
			return errs.InsufficientBalance
		}
		if err = s.budgetHistoryRepo.UpdateBalanceTX(ctx, tx, history.TransactionID.String, newAmount); err != nil {
			zap.L().Sugar().Errorf("Failed to update balance for transactionID=%s: %v", history.TransactionID.String, err)
			return errs.DatabaseError
		}
	}

	zap.L().Sugar().Infof("Successfully updated budget history for budgetID=%s", budgetID)
	return nil
}
