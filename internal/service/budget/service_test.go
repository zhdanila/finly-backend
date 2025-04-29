package budget

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/internal/repository/budget/mock"
	mock_budget_history "finly-backend/internal/repository/budget_history/mock"
	"finly-backend/pkg/transaction"
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
	mockBudgetRepo := mock.NewMockBudget(ctrl)
	mockBudgetHistoryRepo := mock_budget_history.NewMockBudgetHistory(ctrl)

	mockDB := &sqlx.DB{}
	mockTx := &sqlx.Tx{}

	tests := []struct {
		name        string
		req         *CreateBudgetRequest
		mockSetup   func()
		expectedRes *CreateBudgetResponse
		expectedErr error
	}{
		{
			name: "Successful budget creation with initial amount",
			req: &CreateBudgetRequest{
				UserID:   "user123",
				Currency: "USD",
				Amount:   100.00,
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetDB().Return(mockDB)
				mockBudgetRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "USD").
					Return("budget123", nil)
				mockBudgetHistoryRepo.EXPECT().CreateInitialTX(ctx, mockTx, "budget123", 100.00).
					Return("history123", nil)
			},
			expectedRes: &CreateBudgetResponse{ID: "budget123"},
			expectedErr: nil,
		},
		{
			name: "Successful budget creation without initial amount",
			req: &CreateBudgetRequest{
				UserID:   "user123",
				Currency: "USD",
				Amount:   0,
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetDB().Return(mockDB)
				mockBudgetRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "USD").
					Return("budget123", nil)
				// No history call expected
			},
			expectedRes: &CreateBudgetResponse{ID: "budget123"},
			expectedErr: nil,
		},
		{
			name: "CreateTX error",
			req: &CreateBudgetRequest{
				UserID:   "user123",
				Currency: "USD",
				Amount:   50,
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetDB().Return(mockDB)
				mockBudgetRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "USD").
					Return("", errors.New("create error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("create error"),
		},
		{
			name: "CreateInitialTX error",
			req: &CreateBudgetRequest{
				UserID:   "user123",
				Currency: "USD",
				Amount:   100,
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetDB().Return(mockDB)
				mockBudgetRepo.EXPECT().CreateTX(ctx, mockTx, "user123", "USD").
					Return("budget123", nil)
				mockBudgetHistoryRepo.EXPECT().CreateInitialTX(ctx, mockTx, "budget123", 100.00).
					Return("", errors.New("history error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("history error"),
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

			service := NewService(mockBudgetRepo, mockBudgetHistoryRepo, mockTxExec)

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

func TestGetByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgetRepo := mock.NewMockBudget(ctrl)
	mockBudgetHistoryRepo := mock_budget_history.NewMockBudgetHistory(ctrl)
	service := NewService(mockBudgetRepo, mockBudgetHistoryRepo, transaction.NewTransactionExecutor())
	ctx := context.Background()

	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name        string
		req         *GetBudgetByIDRequest
		mockSetup   func()
		expectedRes *GetBudgetByIDResponse
		expectedErr error
	}{
		{
			name: "Successful get budget",
			req: &GetBudgetByIDRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetByUserID(ctx, "user123").
					Return(&domain.Budget{
						ID:        "budget123",
						UserID:    "user123",
						Currency:  "USD",
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					}, nil)
			},
			expectedRes: &GetBudgetByIDResponse{
				BudgetObject: &BudgetObject{
					ID:        "budget123",
					UserID:    "user123",
					Currency:  "USD",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
			},
			expectedErr: nil,
		},
		{
			name: "No budget found",
			req: &GetBudgetByIDRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetByUserID(ctx, "user123").
					Return(nil, sql.ErrNoRows)
			},
			expectedRes: &GetBudgetByIDResponse{},
			expectedErr: nil,
		},
		{
			name: "Error getting budget",
			req: &GetBudgetByIDRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockBudgetRepo.EXPECT().GetByUserID(ctx, "user123").
					Return(nil, errors.New("database error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := service.GetByUserID(ctx, tt.req)

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

func TestGetBudgetHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgetRepo := mock.NewMockBudget(ctrl)
	mockBudgetHistoryRepo := mock_budget_history.NewMockBudgetHistory(ctrl)
	service := NewService(mockBudgetRepo, mockBudgetHistoryRepo, transaction.NewTransactionExecutor())
	ctx := context.Background()

	createdAt := time.Now()

	tests := []struct {
		name        string
		req         *GetBudgetHistoryRequest
		mockSetup   func()
		expectedRes *GetBudgetHistoryResponse
		expectedErr error
	}{
		{
			name: "Successful get budget history",
			req: &GetBudgetHistoryRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().List(ctx, "budget123").
					Return([]*domain.BudgetHistory{
						{
							ID:        "history1",
							BudgetID:  "budget123",
							Balance:   100.00,
							CreatedAt: createdAt,
						},
						{
							ID:        "history2",
							BudgetID:  "budget123",
							Balance:   150.00,
							CreatedAt: createdAt.Add(1 * time.Hour),
						},
					}, nil)
			},
			expectedRes: &GetBudgetHistoryResponse{
				BudgetHistory: []*BudgetHistory{
					{
						ID:        "history1",
						BudgetID:  "budget123",
						Balance:   100.00,
						CreatedAt: createdAt,
					},
					{
						ID:        "history2",
						BudgetID:  "budget123",
						Balance:   150.00,
						CreatedAt: createdAt.Add(1 * time.Hour),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "No budget history found",
			req: &GetBudgetHistoryRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().List(ctx, "budget123").
					Return(nil, sql.ErrNoRows)
			},
			expectedRes: &GetBudgetHistoryResponse{},
			expectedErr: nil,
		},
		{
			name: "Error getting budget history",
			req: &GetBudgetHistoryRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().List(ctx, "budget123").
					Return(nil, errors.New("database error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.mockSetup()

			// Call the service
			resp, err := service.GetBudgetHistory(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.expectedRes.BudgetHistory == nil {
					assert.Empty(t, resp.BudgetHistory)
				} else {
					assert.Equal(t, tt.expectedRes, resp)
				}
			}
		})
	}
}

func TestGetCurrentBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgetRepo := mock.NewMockBudget(ctrl)
	mockBudgetHistoryRepo := mock_budget_history.NewMockBudgetHistory(ctrl)
	service := NewService(mockBudgetRepo, mockBudgetHistoryRepo, transaction.NewTransactionExecutor())
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *GetCurrentBalanceRequest
		mockSetup   func()
		expectedRes *GetCurrentBalanceResponse
		expectedErr error
	}{
		{
			name: "Successful get current balance",
			req: &GetCurrentBalanceRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().GetCurrentBalance(ctx, "budget123").
					Return(150.00, nil)
			},
			expectedRes: &GetCurrentBalanceResponse{
				Balance: 150.00,
			},
			expectedErr: nil,
		},
		{
			name: "No balance found",
			req: &GetCurrentBalanceRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().GetCurrentBalance(ctx, "budget123").
					Return(0.0, sql.ErrNoRows)
			},
			expectedRes: &GetCurrentBalanceResponse{},
			expectedErr: nil,
		},
		{
			name: "Error getting balance",
			req: &GetCurrentBalanceRequest{
				BudgetID: "budget123",
			},
			mockSetup: func() {
				mockBudgetHistoryRepo.EXPECT().GetCurrentBalance(ctx, "budget123").
					Return(0.0, errors.New("database error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.mockSetup()

			// Call the service
			resp, err := service.GetCurrentBalance(ctx, tt.req)

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
