package handler

import (
	"bytes"
	"encoding/json"
	"finly-backend/internal/domain/enums/e_transaction_type"
	"finly-backend/internal/service"
	"finly-backend/internal/service/transaction"
	"finly-backend/internal/service/transaction/mock"
	"finly-backend/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTransactionTest(t *testing.T) (*echo.Echo, *mock.MockTransaction, *Transaction) {
	var err error

	ctrl := gomock.NewController(t)
	mockTransaction := mock.NewMockTransaction(ctrl)
	service := &service.Service{Transaction: mockTransaction}
	handler := NewTransaction(service)
	e := echo.New()

	if e.Validator, err = validator.CustomValidator(); err != nil {
		zap.L().Fatal("Error setting up custom validator", zap.Error(err))
	}

	return e, mockTransaction, handler
}

func TestTransaction_Create(t *testing.T) {
	e, mockTransaction, handler := setupTransactionTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          transaction.CreateTransactionRequest
		userID         string
		mockResponse   *transaction.CreateTransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful transaction creation",
			input: transaction.CreateTransactionRequest{
				UserID:     "user123",
				CategoryID: "category123",
				BudgetID:   "budget123",
				Amount:     100.0,
				Type:       e_transaction_type.Deposit,
				Note:       "Grocery purchase",
			},
			userID:         "user123",
			mockResponse:   &transaction.CreateTransactionResponse{ID: "transaction123"},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid input",
			input: transaction.CreateTransactionRequest{
				UserID:     "",
				CategoryID: "",
				BudgetID:   "",
				Amount:     0,
				Type:       "",
			},
			userID:         "",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/transaction", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockTransaction.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Create(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response transaction.CreateTransactionResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
		})
	}
}

func TestTransaction_List(t *testing.T) {
	e, mockTransaction, handler := setupTransactionTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		userID         string
		input          transaction.ListTransactionRequest
		mockResponse   *transaction.ListTransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful transactions list retrieval",
			userID: "user123",
			input: transaction.ListTransactionRequest{
				UserID: "user123",
			},
			mockResponse: &transaction.ListTransactionResponse{
				Transactions: []transaction.TransactionObject{
					{
						ID:         "transaction123",
						UserID:     "user123",
						CategoryID: "category123",
						BudgetID:   "budget123",
						Amount:     100.0,
						Type:       e_transaction_type.Deposit,
						Note:       "Grocery purchase",
						CreatedAt:  time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid input",
			userID: "",
			input: transaction.ListTransactionRequest{
				UserID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/transaction?user_id="+tt.userID, nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockTransaction.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.List(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response transaction.ListTransactionResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockResponse.Transactions), len(response.Transactions))
				if len(tt.mockResponse.Transactions) > 0 {
					assert.Equal(t, tt.mockResponse.Transactions[0].ID, response.Transactions[0].ID)
				}
			}
		})
	}
}

func TestTransaction_Update(t *testing.T) {
	e, mockTransaction, handler := setupTransactionTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		transactionID  string
		userID         string
		input          transaction.UpdateTransactionRequest
		mockResponse   *transaction.UpdateTransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:          "successful transaction update",
			transactionID: "transaction123",
			userID:        "user123",
			input: transaction.UpdateTransactionRequest{
				UserID:        "user123",
				TransactionID: "transaction123",
				CategoryID:    "category456",
				BudgetID:      456,
				Amount:        200.0,
				Type:          string(e_transaction_type.Withdrawal),
				Note:          "Updated note",
			},
			mockResponse:   &transaction.UpdateTransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:          "invalid input",
			transactionID: "",
			userID:        "",
			input: transaction.UpdateTransactionRequest{
				UserID:        "",
				TransactionID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPatch, "/transaction/"+tt.transactionID, bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.transactionID)

			if tt.mockResponse != nil {
				mockTransaction.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Update(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response transaction.UpdateTransactionResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransaction_Delete(t *testing.T) {
	e, mockTransaction, handler := setupTransactionTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		transactionID  string
		userID         string
		input          transaction.DeleteTransactionRequest
		mockResponse   *transaction.DeleteTransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:          "successful transaction deletion",
			transactionID: "transaction123",
			userID:        "user123",
			input: transaction.DeleteTransactionRequest{
				UserID:        "user123",
				TransactionID: "transaction123",
			},
			mockResponse:   &transaction.DeleteTransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:          "invalid input",
			transactionID: "",
			userID:        "",
			input: transaction.DeleteTransactionRequest{
				UserID:        "",
				TransactionID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/transaction/"+tt.transactionID, nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.transactionID)

			if tt.mockResponse != nil {
				mockTransaction.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Delete(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response transaction.DeleteTransactionResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}
