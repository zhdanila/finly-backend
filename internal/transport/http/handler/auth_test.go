package handler

import (
	"bytes"
	"encoding/json"
	"finly-backend/internal/service"
	"finly-backend/internal/service/auth"
	"finly-backend/internal/service/auth/mock"
	validator2 "finly-backend/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupAuthTest(t *testing.T) (*echo.Echo, *mock.MockAuth, *Auth) {
	var err error

	ctrl := gomock.NewController(t)
	mockAuth := mock.NewMockAuth(ctrl)
	service := &service.Service{Auth: mockAuth}
	handler := NewAuth(service)
	e := echo.New()

	if e.Validator, err = validator2.CustomValidator(); err != nil {
		zap.L().Fatal("Error setting up custom validator", zap.Error(err))
	}

	return e, mockAuth, handler
}

func TestAuth_RegisterUser(t *testing.T) {
	e, mockAuth, handler := setupAuthTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          auth.RegisterRequest
		mockResponse   *auth.RegisterResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful registration",
			input: auth.RegisterRequest{
				UserInfo: auth.UserInfo{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
				},
				Password: "password123",
			},
			mockResponse:   &auth.RegisterResponse{Token: "jwt_token"},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid input",
			input: auth.RegisterRequest{
				UserInfo: auth.UserInfo{
					FirstName: "",
					LastName:  "Doe",
					Email:     "invalid_email",
				},
				Password: "pass",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockAuth.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.RegisterUser(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response auth.RegisterResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	e, mockAuth, handler := setupAuthTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          auth.LoginRequest
		mockResponse   *auth.LoginResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful login",
			input: auth.LoginRequest{
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			mockResponse:   &auth.LoginResponse{Token: "jwt_token"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid input",
			input: auth.LoginRequest{
				Email:    "invalid_email",
				Password: "pass",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockAuth.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Login(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response auth.LoginResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}
		})
	}
}

func TestAuth_Logout(t *testing.T) {
	e, mockAuth, handler := setupAuthTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          auth.LogoutRequest
		mockResponse   *auth.LogoutResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful logout",
			input:          auth.LogoutRequest{AuthToken: "Bearer jwt_token"},
			mockResponse:   &auth.LogoutResponse{Message: "Logged out successfully"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			input:          auth.LogoutRequest{AuthToken: ""},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			req.Header.Set("Authorization", tt.input.AuthToken)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockAuth.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Logout(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response auth.LogoutResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Message, response.Message)
			}
		})
	}
}

func TestAuth_Refresh(t *testing.T) {
	e, mockAuth, handler := setupAuthTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          auth.RefreshTokenRequest
		mockResponse   *auth.RefreshTokenResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful refresh",
			input:          auth.RefreshTokenRequest{AuthToken: "Bearer jwt_token"},
			mockResponse:   &auth.RefreshTokenResponse{Token: "new_jwt_token"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			input:          auth.RefreshTokenRequest{AuthToken: ""},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			req.Header.Set("Authorization", tt.input.AuthToken)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockAuth.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Refresh(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response auth.RefreshTokenResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}
		})
	}
}

func TestAuth_Me(t *testing.T) {
	e, mockAuth, handler := setupAuthTest(t)
	defer gomock.NewController(t).Finish()

	tests := []struct {
		name           string
		input          auth.MeRequest
		mockResponse   *auth.MeResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:  "successful user info retrieval",
			input: auth.MeRequest{AuthToken: "Bearer jwt_token"},
			mockResponse: &auth.MeResponse{
				UserInfo: auth.UserInfo{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			input:          auth.MeRequest{AuthToken: ""},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			req.Header.Set("Authorization", tt.input.AuthToken)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.mockResponse != nil {
				mockAuth.EXPECT().
					Me(gomock.Any(), gomock.Any()).
					Return(tt.mockResponse, tt.mockError)
			}

			err := handler.Me(c)

			if tt.expectedStatus == http.StatusBadRequest {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.mockResponse != nil {
				var response auth.MeResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.UserInfo, response.UserInfo)
			}
		})
	}
}
