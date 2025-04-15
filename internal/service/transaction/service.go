package transaction

import (
	"context"
	"finly-backend/internal/domain"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"finly-backend/internal/repository"
	"strings"
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

func (s *Service) Create(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	var err error

	tx, err := s.transactionRepo.GetDB().BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	transactionID, err := s.transactionRepo.CreateTX(ctx, tx, req.UserID, req.BudgetID, req.CategoryID, req.Type.String(), req.Note, req.Amount)
	if err != nil {
		return nil, err
	}

	lastBudgetHistory, err := s.budgetHistoryRepo.GetLastByID(ctx, req.BudgetID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			lastBudgetHistory = nil
		} else {
			return nil, err
		}
	}

	newAmount, err := calculateNewAmount(lastBudgetHistory, req.Amount, req.Type.String())
	if err != nil {
		return nil, err
	}

	if _, err = s.budgetHistoryRepo.CreateTX(ctx, tx, req.BudgetID, newAmount); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateTransactionResponse{
		ID: transactionID,
	}, nil
}

func calculateNewAmount(budgetHistory *domain.BudgetHistory, amount float64, transactionType string) (float64, error) {
	var newAmount float64
	switch transactionType {
	case e_transaction_type.Deposit.String():
		if budgetHistory == nil {
			newAmount = amount
		} else {
			newAmount = budgetHistory.Balance + amount
		}
	case e_transaction_type.Withdrawal.String():
		if budgetHistory == nil || budgetHistory.Balance < amount {
			return 0, errs.InsufficientBalance
		}
		newAmount = budgetHistory.Balance - amount
	default:
		return 0, errs.InvalidTransactionType
	}
	return newAmount, nil
}

func (s *Service) List(ctx context.Context, req *ListTransactionRequest) (*ListTransactionResponse, error) {
	var err error

	transactions, err := s.transactionRepo.List(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	var transactionList []Transaction
	for _, transaction := range transactions {
		transactionList = append(transactionList, Transaction{
			ID:         transaction.ID,
			UserID:     transaction.UserID,
			BudgetID:   transaction.BudgetID,
			CategoryID: transaction.CategoryID,
			Type:       e_transaction_type.Enum(transaction.TransactionType),
			Note:       transaction.Note,
			Amount:     transaction.Amount,
			CreatedAt:  transaction.CreatedAt,
		})
	}

	return &ListTransactionResponse{
		Transactions: transactionList,
	}, nil
}

func (s *Service) Update(ctx context.Context, req *UpdateTransactionRequest) (*UpdateTransactionResponse, error) {
	var err error

	if err = s.transactionRepo.Update(ctx, req.TransactionID, req.UserID, req.CategoryID, req.Type, req.Note, req.Amount); err != nil {
		return nil, err
	}

	return &UpdateTransactionResponse{}, nil
}
