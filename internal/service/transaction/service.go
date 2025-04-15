package transaction

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"finly-backend/internal/repository"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type Service struct {
	transactionRepo   repository.Transaction
	budgetHistoryRepo repository.BudgetHistory
}

func NewService(transactionRepo repository.Transaction, budgetHistoryRepo repository.BudgetHistory) *Service {
	return &Service{
		transactionRepo:   transactionRepo,
		budgetHistoryRepo: budgetHistoryRepo,
	}
}

// withTransaction manages database transactions.
func (s *Service) withTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.transactionRepo.GetDB().BeginTxx(ctx, nil)
	if err != nil {
		return errs.DatabaseError
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v; original error: %v", rbErr, err)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errs.DatabaseError
	}
	return nil
}

func (s *Service) Create(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	var transactionID string
	if err := s.withTransaction(ctx, func(tx *sqlx.Tx) error {
		var err error
		transactionID, err = s.transactionRepo.CreateTX(ctx, tx, req.UserID, req.BudgetID, req.CategoryID, req.Type.String(), req.Note, req.Amount)
		if err != nil {
			return errs.DatabaseError
		}

		lastBudgetHistory, err := s.budgetHistoryRepo.GetLastByBudgetID(ctx, req.BudgetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				lastBudgetHistory = nil
			} else {
				return errs.DatabaseError
			}
		}

		newAmount, err := calculateNewAmount(lastBudgetHistory, req.Amount, req.Type.String())
		if err != nil {
			return errs.InvalidInput
		}

		_, err = s.budgetHistoryRepo.CreateTX(ctx, tx, req.BudgetID, transactionID, newAmount)
		if err != nil {
			return errs.DatabaseError
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &CreateTransactionResponse{ID: transactionID}, nil
}

func (s *Service) List(ctx context.Context, req *ListTransactionRequest) (*ListTransactionResponse, error) {
	transactions, err := s.transactionRepo.List(ctx, req.UserID)
	if err != nil {
		return nil, errs.DatabaseError
	}

	transactionList := make([]Transaction, 0, len(transactions))
	for _, t := range transactions {
		transactionList = append(transactionList, Transaction{
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
	if err := s.withTransaction(ctx, func(tx *sqlx.Tx) error {
		if err := s.transactionRepo.UpdateTX(ctx, tx, req.TransactionID, req.UserID, req.CategoryID, req.Type, req.Note, req.Amount); err != nil {
			return errs.DatabaseError
		}

		transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID, req.UserID)
		if err != nil {
			return errs.DatabaseError
		}

		if transaction.TransactionType != req.Type || transaction.Amount != req.Amount {
			difference, err := calculateDifferenceBetweenTransactions(transaction.TransactionType, transaction.Amount, req.Type, req.Amount)
			if err != nil {
				return errs.InvalidTransactionType
			}

			if err := s.updateBudgetHistory(ctx, tx, transaction.BudgetID, transaction.CreatedAt, difference, true); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &UpdateTransactionResponse{}, nil
}

func (s *Service) Delete(ctx context.Context, req *DeleteTransactionRequest) (*DeleteTransactionResponse, error) {
	if err := s.withTransaction(ctx, func(tx *sqlx.Tx) error {
		transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID, req.UserID)
		if err != nil {
			return errs.DatabaseError
		}

		difference, err := calculateDifference(transaction.TransactionType, transaction.Amount, false)
		if err != nil {
			return errs.InvalidTransactionType
		}

		if err = s.updateBudgetHistory(ctx, tx, transaction.BudgetID, transaction.CreatedAt, difference, false); err != nil {
			return err
		}

		if err = s.transactionRepo.DeleteTX(ctx, tx, req.TransactionID, req.UserID); err != nil {
			return errs.DatabaseError
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &DeleteTransactionResponse{}, nil
}

func (s *Service) updateBudgetHistory(ctx context.Context, tx *sqlx.Tx, budgetID string, fromDate time.Time, difference float64, inclusiveDate bool) error {
	budgetHistory, err := s.budgetHistoryRepo.ListFromDate(ctx, budgetID, fromDate, inclusiveDate)
	if err != nil {
		return errs.DatabaseError
	}

	for _, history := range budgetHistory {
		newAmount := history.Balance + difference
		if newAmount < 0 {
			return errs.InsufficientBalance
		}
		if err := s.budgetHistoryRepo.UpdateBalanceTX(ctx, tx, history.TransactionID.String, newAmount); err != nil {
			return errs.DatabaseError
		}
	}
	return nil
}
