package transaction

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"finly-backend/internal/repository/budget_history/mock"
	mock_transaction "finly-backend/internal/repository/transaction/mock"
	transactionExec "finly-backend/pkg/transaction"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type mockTransactionExecutor struct {
	withTx func(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error
}

func (m *mockTransactionExecutor) WithTransaction(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	return m.withTx(ctx, db, fn)
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTransactionRepo := mock_transaction.NewMockTransaction(ctrl)
	mockBudgetHistoryRepo := mock.NewMockBudgetHistory(ctrl)
	mockDB := &sqlx.DB{}
	mockTx := &sqlx.Tx{}

	tests := []struct {
		name        string
		req         *CreateTransactionRequest
		mockSetup   func()
		expectedRes *CreateTransactionResponse
		expectedErr error
	}{
		{
			name: "Successful transaction creation with deposit",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Deposit,
				Note:       "Test deposit",
				Amount:     100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "deposit", "Test deposit", 100.00).
					Return("trans123", nil)
				mockBudgetHistoryRepo.EXPECT().GetLastByBudgetID(ctx, "budget123").
					Return(&domain.BudgetHistory{Balance: 50.00}, nil)
				mockBudgetHistoryRepo.EXPECT().CreateTX(ctx, mockTx, "budget123", "trans123", 150.00).
					Return("history123", nil)
			},
			expectedRes: &CreateTransactionResponse{ID: "trans123"},
			expectedErr: nil,
		},
		{
			name: "Successful transaction creation with withdrawal",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Withdrawal,
				Note:       "Test withdrawal",
				Amount:     50.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "withdrawal", "Test withdrawal", 50.00).
					Return("trans123", nil)
				mockBudgetHistoryRepo.EXPECT().GetLastByBudgetID(ctx, "budget123").
					Return(&domain.BudgetHistory{Balance: 100.00}, nil)
				mockBudgetHistoryRepo.EXPECT().CreateTX(ctx, mockTx, "budget123", "trans123", 50.00).
					Return("history123", nil)
			},
			expectedRes: &CreateTransactionResponse{ID: "trans123"},
			expectedErr: nil,
		},
		{
			name: "No previous budget history",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Deposit,
				Note:       "Test deposit",
				Amount:     100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "deposit", "Test deposit", 100.00).
					Return("trans123", nil)
				mockBudgetHistoryRepo.EXPECT().GetLastByBudgetID(ctx, "budget123").
					Return(nil, sql.ErrNoRows)
				mockBudgetHistoryRepo.EXPECT().CreateTX(ctx, mockTx, "budget123", "trans123", 100.00).
					Return("history123", nil)
			},
			expectedRes: &CreateTransactionResponse{ID: "trans123"},
			expectedErr: nil,
		},
		{
			name: "CreateTX error",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Deposit,
				Amount:     100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "deposit", "", 100.00).
					Return("", errors.New("create error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "GetLastByBudgetID error",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Deposit,
				Amount:     100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "deposit", "", 100.00).
					Return("trans123", nil)
				mockBudgetHistoryRepo.EXPECT().GetLastByBudgetID(ctx, "budget123").
					Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "CreateTX budget history error",
			req: &CreateTransactionRequest{
				UserID:     "user123",
				BudgetID:   "budget123",
				CategoryID: "cat123",
				Type:       e_transaction_type.Deposit,
				Amount:     100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "budget123", "cat123", "deposit", "", 100.00).
					Return("trans123", nil)
				mockBudgetHistoryRepo.EXPECT().GetLastByBudgetID(ctx, "budget123").
					Return(&domain.BudgetHistory{Balance: 50.00}, nil)
				mockBudgetHistoryRepo.EXPECT().CreateTX(ctx, mockTx, "budget123", "trans123", 150.00).
					Return("", errors.New("history error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			mockTxExec := &mockTransactionExecutor{
				withTx: func(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
					return fn(mockTx)
				},
			}

			service := NewService(mockTransactionRepo, mockBudgetHistoryRepo, mockTxExec)

			resp, err := service.Create(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTransactionRepo := mock_transaction.NewMockTransaction(ctrl)
	mockBudgetHistoryRepo := mock.NewMockBudgetHistory(ctrl)
	service := NewService(mockTransactionRepo, mockBudgetHistoryRepo, transactionExec.NewTransactionExecutor())

	createdAt := time.Now()

	tests := []struct {
		name        string
		req         *ListTransactionRequest
		mockSetup   func()
		expectedRes *ListTransactionResponse
		expectedErr error
	}{
		{
			name: "Successful list transactions",
			req: &ListTransactionRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().List(ctx, "user123").Return([]*domain.Transaction{
					{
						ID:              "trans1",
						UserID:          "user123",
						BudgetID:        "budget123",
						CategoryID:      "cat123",
						TransactionType: "deposit",
						Note:            "Test deposit",
						Amount:          100.00,
						CreatedAt:       createdAt,
					},
					{
						ID:              "trans2",
						UserID:          "user123",
						BudgetID:        "budget123",
						CategoryID:      "cat123",
						TransactionType: "withdrawal",
						Note:            "Test withdrawal",
						Amount:          50.00,
						CreatedAt:       createdAt.Add(time.Hour),
					},
				}, nil)
			},
			expectedRes: &ListTransactionResponse{
				Transactions: []TransactionObject{
					{
						ID:         "trans1",
						UserID:     "user123",
						BudgetID:   "budget123",
						CategoryID: "cat123",
						Type:       e_transaction_type.Deposit,
						Note:       "Test deposit",
						Amount:     100.00,
						CreatedAt:  createdAt,
					},
					{
						ID:         "trans2",
						UserID:     "user123",
						BudgetID:   "budget123",
						CategoryID: "cat123",
						Type:       e_transaction_type.Withdrawal,
						Note:       "Test withdrawal",
						Amount:     50.00,
						CreatedAt:  createdAt.Add(time.Hour),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "List error",
			req: &ListTransactionRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().List(ctx, "user123").Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := service.List(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTransactionRepo := mock_transaction.NewMockTransaction(ctrl)
	mockBudgetHistoryRepo := mock.NewMockBudgetHistory(ctrl)
	mockDB := &sqlx.DB{}
	mockTx := &sqlx.Tx{}

	tests := []struct {
		name        string
		req         *UpdateTransactionRequest
		mockSetup   func()
		expectedRes *UpdateTransactionResponse
		expectedErr error
	}{
		{
			name: "Successful update with type and amount change",
			req: &UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
				CategoryID:    "cat123",
				Type:          "withdrawal",
				Note:          "Updated note",
				Amount:        50.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().UpdateTX(ctx, mockTx, "trans123", "user123", "cat123", "withdrawal", "Updated note", 50.00).
					Return(nil)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          100.00,
						CreatedAt:       time.Now(),
					}, nil)
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", gomock.Any(), true).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
					}, nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history1", 50.00).
					Return(nil)
			},
			expectedRes: &UpdateTransactionResponse{},
			expectedErr: nil,
		},
		{
			name: "Successful update without type or amount change",
			req: &UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
				CategoryID:    "cat123",
				Type:          "deposit",
				Note:          "Updated note",
				Amount:        100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().UpdateTX(ctx, mockTx, "trans123", "user123", "cat123", "deposit", "Updated note", 100.00).
					Return(nil)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          100.00,
						CreatedAt:       time.Now(),
					}, nil)
			},
			expectedRes: &UpdateTransactionResponse{},
			expectedErr: nil,
		},
		{
			name: "UpdateTX error",
			req: &UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
				CategoryID:    "cat123",
				Type:          "deposit",
				Amount:        100.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().UpdateTX(ctx, mockTx, "trans123", "user123", "cat123", "deposit", "", 100.00).
					Return(errors.New("update error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "GetByID error",
			req: &UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
				CategoryID:    "cat123",
				Type:          "withdrawal",
				Amount:        50.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().UpdateTX(ctx, mockTx, "trans123", "user123", "cat123", "withdrawal", "", 50.00).
					Return(nil)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(nil, errors.New("get error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "Insufficient balance",
			req: &UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
				CategoryID:    "cat123",
				Type:          "withdrawal",
				Amount:        300.00,
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().UpdateTX(ctx, mockTx, "trans123", "user123", "cat123", "withdrawal", "", 300.00).
					Return(nil)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          100.00,
						CreatedAt:       time.Now(),
					}, nil)
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", gomock.Any(), true).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
					}, nil)
				// Delta: -100 (revert deposit) - 300 (new withdrawal) = -400, new balance = 200 - 400 = -200 (insufficient)
			},
			expectedRes: nil,
			expectedErr: errs.InsufficientBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			mockTxExec := &mockTransactionExecutor{
				withTx: func(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
					return fn(mockTx)
				},
			}

			service := NewService(mockTransactionRepo, mockBudgetHistoryRepo, mockTxExec)

			resp, err := service.Update(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTransactionRepo := mock_transaction.NewMockTransaction(ctrl)
	mockBudgetHistoryRepo := mock.NewMockBudgetHistory(ctrl)
	mockDB := &sqlx.DB{}
	mockTx := &sqlx.Tx{}

	tests := []struct {
		name        string
		req         *DeleteTransactionRequest
		mockSetup   func()
		expectedRes *DeleteTransactionResponse
		expectedErr error
	}{
		{
			name: "Successful delete with deposit",
			req: &DeleteTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          100.00,
						CreatedAt:       time.Now(),
					}, nil)
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", gomock.Any(), false).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
					}, nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history1", 100.00).
					Return(nil)
				mockTransactionRepo.EXPECT().DeleteTX(ctx, mockTx, "trans123", "user123").
					Return(nil)
			},
			expectedRes: &DeleteTransactionResponse{},
			expectedErr: nil,
		},
		{
			name: "GetByID error",
			req: &DeleteTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(nil, errors.New("get error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "DeleteTX error",
			req: &DeleteTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          100.00,
						CreatedAt:       time.Now(),
					}, nil)
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", gomock.Any(), false).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
					}, nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history1", 100.00).
					Return(nil)
				mockTransactionRepo.EXPECT().DeleteTX(ctx, mockTx, "trans123", "user123").
					Return(errors.New("delete error"))
			},
			expectedRes: nil,
			expectedErr: errs.DatabaseError,
		},
		{
			name: "Insufficient balance after deletion",
			req: &DeleteTransactionRequest{
				UserID:        "user123",
				TransactionID: "trans123",
			},
			mockSetup: func() {
				mockTransactionRepo.EXPECT().GetDB().Return(mockDB)
				mockTransactionRepo.EXPECT().GetByID(ctx, "trans123", "user123").
					Return(&domain.Transaction{
						ID:              "trans123",
						UserID:          "user123",
						BudgetID:        "budget123",
						TransactionType: "deposit",
						Amount:          300.00,
						CreatedAt:       time.Now(),
					}, nil)
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", gomock.Any(), false).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
					}, nil)
			},
			expectedRes: nil,
			expectedErr: errs.InsufficientBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			mockTxExec := &mockTransactionExecutor{
				withTx: func(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
					return fn(mockTx)
				},
			}

			service := NewService(mockTransactionRepo, mockBudgetHistoryRepo, mockTxExec)

			resp, err := service.Delete(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestUpdateBudgetHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTransactionRepo := mock_transaction.NewMockTransaction(ctrl)
	mockBudgetHistoryRepo := mock.NewMockBudgetHistory(ctrl)
	mockTx := &sqlx.Tx{}
	fromDate := time.Now()

	tests := []struct {
		name        string
		budgetID    string
		fromDate    time.Time
		difference  float64
		inclusive   bool
		mockSetup   func()
		expectedErr error
	}{
		{
			name:       "Successful update budget history",
			budgetID:   "budget123",
			fromDate:   fromDate,
			difference: -50.00,
			inclusive:  true,
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", fromDate, true).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 200.00},
						{TransactionID: sql.NullString{String: "history2", Valid: true}, Balance: 150.00},
					}, nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history1", 150.00).
					Return(nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history2", 100.00).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:       "Insufficient balance",
			budgetID:   "budget123",
			fromDate:   fromDate,
			difference: -200.00,
			inclusive:  true,
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", fromDate, true).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 100.00},
					}, nil)
			},
			expectedErr: errs.InsufficientBalance,
		},
		{
			name:       "ListFromDate error",
			budgetID:   "budget123",
			fromDate:   fromDate,
			difference: 50.00,
			inclusive:  true,
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", fromDate, true).
					Return(nil, errors.New("db error"))
			},
			expectedErr: errs.DatabaseError,
		},
		{
			name:       "UpdateBalanceTX error",
			budgetID:   "budget123",
			fromDate:   fromDate,
			difference: 50.00,
			inclusive:  true,
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().ListFromDate(ctx, "budget123", fromDate, true).
					Return([]*domain.BudgetHistory{
						{TransactionID: sql.NullString{String: "history1", Valid: true}, Balance: 100.00},
					}, nil)
				mockBudgetHistoryRepo.EXPECT().UpdateBalanceTX(ctx, mockTx, "history1", 150.00).
					Return(errors.New("update error"))
			},
			expectedErr: errs.DatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockTransactionRepo, mockBudgetHistoryRepo, transactionExec.NewTransactionExecutor())

			err := service.updateBudgetHistory(ctx, mockTx, tt.budgetID, tt.fromDate, tt.difference, tt.inclusive)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
