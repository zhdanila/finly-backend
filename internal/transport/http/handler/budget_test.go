package handler

import (
	"bytes"
	"encoding/json"
	"finly-backend/internal/service"
	"finly-backend/internal/service/budget"
	"finly-backend/internal/service/budget/mock"
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

func setupBudgetTest(t *testing.T) (*echo.Echo, *mock.MockBudget, *Budget) {
	var err error

	ctrl := gomock.NewController(t)
	mockBudget := mock.NewMockBudget(ctrl)
	service := &service.Service{Budget: mockBudget}
	handler := NewBudget(service)
	e := echo.New()

	if e.Validator, err = validator.CustomValidator(); err != nil {
		zap.L().Fatal("Error setting up custom validator", zap.Error(err))
	}

	return e, mockBudget, handler
}

func TestBudget_Create(t *testing.T) {
	e, mockBudget, handler := setupBudgetTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          budget.CreateBudgetRequest
		userID         string
		mockResponse   *budget.CreateBudgetResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful budget creation",
			input: budget.CreateBudgetRequest{
				UserID:   "user123",
				Currency: "USD",
				Amount:   1000.0,
			},
			userID:         "user123",
			mockResponse:   &budget.CreateBudgetResponse{ID: "budget123"},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid input",
			input: budget.CreateBudgetRequest{
				UserID:   "",
				Currency: "",
				Amount:   0,
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
			req := httptest.NewRequest(http.MethodPost, "/budget", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockBudget.EXPECT().
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
				var response budget.CreateBudgetResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
		})
	}
}

func TestBudget_GetByUserID(t *testing.T) {
	e, mockBudget, handler := setupBudgetTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          budget.GetBudgetByIDRequest
		userID         string
		mockResponse   *budget.GetBudgetByIDResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful budget retrieval",
			input: budget.GetBudgetByIDRequest{
				UserID: "user123",
			},
			userID: "user123",
			mockResponse: &budget.GetBudgetByIDResponse{
				BudgetObject: &budget.BudgetObject{
					ID:        "budget123",
					UserID:    "user123",
					Currency:  "USD",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid input",
			input: budget.GetBudgetByIDRequest{
				UserID: "",
			},
			userID:         "",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/budget", nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockBudget.EXPECT().
					GetByUserID(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.GetByUserID(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response budget.GetBudgetByIDResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.BudgetObject.ID, response.BudgetObject.ID)
			}
		})
	}
}

func TestBudget_GetBudgetHistory(t *testing.T) {
	e, mockBudget, handler := setupBudgetTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		budgetID       string
		input          budget.GetBudgetHistoryRequest
		mockResponse   *budget.GetBudgetHistoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:     "successful budget history retrieval",
			budgetID: "budget123",
			input: budget.GetBudgetHistoryRequest{
				BudgetID: "budget123",
			},
			mockResponse: &budget.GetBudgetHistoryResponse{
				BudgetHistory: []*budget.BudgetHistory{
					{
						ID:        "history1",
						BudgetID:  "budget123",
						Balance:   1000.0,
						CreatedAt: time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid budget ID",
			budgetID: "",
			input: budget.GetBudgetHistoryRequest{
				BudgetID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/budget/"+tt.budgetID+"/history", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("budget_id")
			c.SetParamValues(tt.budgetID)

			if tt.mockResponse != nil {
				mockBudget.EXPECT().
					GetBudgetHistory(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.GetBudgetHistory(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response budget.GetBudgetHistoryResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockResponse.BudgetHistory), len(response.BudgetHistory))
				if len(tt.mockResponse.BudgetHistory) > 0 {
					assert.Equal(t, tt.mockResponse.BudgetHistory[0].ID, response.BudgetHistory[0].ID)
				}
			}
		})
	}
}

func TestBudget_GetCurrentBalance(t *testing.T) {
	e, mockBudget, handler := setupBudgetTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		budgetID       string
		input          budget.GetCurrentBalanceRequest
		mockResponse   *budget.GetCurrentBalanceResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:     "successful balance retrieval",
			budgetID: "budget123",
			input: budget.GetCurrentBalanceRequest{
				BudgetID: "budget123",
			},
			mockResponse: &budget.GetCurrentBalanceResponse{
				Balance: 1000.0,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid budget ID",
			budgetID: "",
			input: budget.GetCurrentBalanceRequest{
				BudgetID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/budget/"+tt.budgetID+"/balance", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("budget_id")
			c.SetParamValues(tt.budgetID)

			if tt.mockResponse != nil {
				mockBudget.EXPECT().
					GetCurrentBalance(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.GetCurrentBalance(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response budget.GetCurrentBalanceResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Balance, response.Balance)
			}
		})
	}
}
