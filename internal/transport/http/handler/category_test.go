package handler

import (
	"bytes"
	"encoding/json"
	"finly-backend/internal/service"
	"finly-backend/internal/service/category"
	"finly-backend/internal/service/category/mock"
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

func setupCategoryTest(t *testing.T) (*echo.Echo, *mock.MockCategory, *Category) {
	var err error

	ctrl := gomock.NewController(t)
	mockCategory := mock.NewMockCategory(ctrl)
	service := &service.Service{Category: mockCategory}
	handler := NewCategory(service)
	e := echo.New()

	if e.Validator, err = validator.CustomValidator(); err != nil {
		zap.L().Fatal("Error setting up custom validator", zap.Error(err))
	}

	return e, mockCategory, handler
}

func TestCategory_Create(t *testing.T) {
	e, mockCategory, handler := setupCategoryTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          category.CreateCategoryRequest
		userID         string
		mockResponse   *category.CreateCategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful category creation",
			input: category.CreateCategoryRequest{
				UserID: "user123",
				CategoryObject: category.CategoryObject{
					Name:           "Groceries",
					IsUserCategory: true,
				},
			},
			userID:         "user123",
			mockResponse:   &category.CreateCategoryResponse{Id: "category123"},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid input",
			input: category.CreateCategoryRequest{
				UserID: "",
				CategoryObject: category.CategoryObject{
					Name: "",
				},
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
			req := httptest.NewRequest(http.MethodPost, "/category", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockCategory.EXPECT().
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
				var response category.CreateCategoryResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Id, response.Id)
			}
		})
	}
}

func TestCategory_GetByID(t *testing.T) {
	e, mockCategory, handler := setupCategoryTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		categoryID     string
		userID         string
		input          category.GetCategoryByIDRequest
		mockResponse   *category.GetCategoryByIDResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:       "successful category retrieval",
			categoryID: "category123",
			userID:     "user123",
			input: category.GetCategoryByIDRequest{
				UserID: "user123",
				ID:     "category123",
			},
			mockResponse: &category.GetCategoryByIDResponse{
				CategoryObject: &category.CategoryObject{
					ID:             "category123",
					UserID:         "user123",
					Name:           "Groceries",
					IsUserCategory: true,
					CreatedAt:      time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid input",
			categoryID: "",
			userID:     "",
			input: category.GetCategoryByIDRequest{
				UserID: "",
				ID:     "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/category/"+tt.categoryID, nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.categoryID)

			if tt.mockResponse != nil {
				mockCategory.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.GetByID(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response category.GetCategoryByIDResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.CategoryObject.ID, response.CategoryObject.ID)
			}
		})
	}
}

func TestCategory_List(t *testing.T) {
	e, mockCategory, handler := setupCategoryTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		userID         string
		input          category.ListCategoriesRequest
		mockResponse   *category.ListCategoriesResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful categories list retrieval",
			userID: "user123",
			input: category.ListCategoriesRequest{
				UserID: "user123",
			},
			mockResponse: &category.ListCategoriesResponse{
				Categories: []category.CategoryObject{
					{
						ID:             "category123",
						UserID:         "user123",
						Name:           "Groceries",
						IsUserCategory: true,
						CreatedAt:      time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid input",
			userID: "",
			input: category.ListCategoriesRequest{
				UserID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/category", nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockCategory.EXPECT().
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
				var response category.ListCategoriesResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockResponse.Categories), len(response.Categories))
				if len(tt.mockResponse.Categories) > 0 {
					assert.Equal(t, tt.mockResponse.Categories[0].ID, response.Categories[0].ID)
				}
			}
		})
	}
}

func TestCategory_ListCustom(t *testing.T) {
	e, mockCategory, handler := setupCategoryTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		userID         string
		input          category.ListCustomCategoriesRequest
		mockResponse   *category.ListCustomCategoriesResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful custom categories list retrieval",
			userID: "user123",
			input: category.ListCustomCategoriesRequest{
				UserID: "user123",
			},
			mockResponse: &category.ListCustomCategoriesResponse{
				Categories: []category.CategoryObject{
					{
						ID:             "category123",
						UserID:         "user123",
						Name:           "Custom Category",
						IsUserCategory: true,
						CreatedAt:      time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid input",
			userID: "",
			input: category.ListCustomCategoriesRequest{
				UserID: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/category/custom", nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockCategory.EXPECT().
					ListCustom(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.ListCustom(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response category.ListCustomCategoriesResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockResponse.Categories), len(response.Categories))
				if len(tt.mockResponse.Categories) > 0 {
					assert.Equal(t, tt.mockResponse.Categories[0].ID, response.Categories[0].ID)
				}
			}
		})
	}
}

func TestCategory_Delete(t *testing.T) {
	e, mockCategory, handler := setupCategoryTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		categoryID     string
		userID         string
		input          category.DeleteCategoryRequest
		mockResponse   *category.DeleteCategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:       "successful category deletion",
			categoryID: "category123",
			userID:     "user123",
			input: category.DeleteCategoryRequest{
				UserID: "user123",
				ID:     "category123",
			},
			mockResponse:   &category.DeleteCategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid input",
			categoryID: "",
			userID:     "",
			input: category.DeleteCategoryRequest{
				UserID: "",
				ID:     "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/category/"+tt.categoryID, nil)
			req.Header.Set("User-Id", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.categoryID)

			if tt.mockResponse != nil {
				mockCategory.EXPECT().
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
				var response category.DeleteCategoryResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}
