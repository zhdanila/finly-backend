package auth

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/internal/repository/auth/mock"
	mock2 "finly-backend/internal/repository/budget/mock"
	"finly-backend/pkg/security"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuth(ctrl)
	mockBudgetRepo := mock2.NewMockBudget(ctrl)
	service := NewService(mockAuthRepo, mockBudgetRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *RegisterRequest
		mockSetup   func()
		expectedErr error
	}{
		{
			name: "Successful registration",
			req: &RegisterRequest{
				UserInfo: UserInfo{
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
				},
				Password: "password123",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().Register(ctx, "test@example.com", gomock.Any(), "John", "Doe").
					Return("user123", nil)
			},
			expectedErr: nil,
		},
		{
			name: "User already exists",
			req: &RegisterRequest{
				UserInfo: UserInfo{
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
				},
				Password: "password123",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().Register(ctx, "test@example.com", gomock.Any(), "John", "Doe").
					Return("", errors.New("duplicate key value violates unique constraint"))
			},
			expectedErr: errs.UserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.mockSetup()

			// Call the service Register method
			resp, err := service.Register(ctx, tt.req)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				// Verify the token is valid by parsing it
				claims, err := security.GetUserFromToken(resp.Token)
				assert.NoError(t, err)
				assert.Equal(t, "user123", claims.UserID)
				assert.Equal(t, "test@example.com", claims.Email)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuth(ctrl)
	mockBudgetRepo := mock2.NewMockBudget(ctrl)
	service := NewService(mockAuthRepo, mockBudgetRepo)
	ctx := context.Background()

	tests := []struct {
		name         string
		req          *LoginRequest
		mockSetup    func()
		expectedResp *LoginResponse
		expectedErr  error
	}{
		{
			name: "Successful login",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				hashedPassword, _ := security.HashPassword("password123")
				mockAuthRepo.EXPECT().GetUserByEmail(ctx, "test@example.com").
					Return(&domain.User{
						ID:           "user123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)
			},
			expectedResp: &LoginResponse{Token: "mocked_jwt_token"},
			expectedErr:  nil,
		},
		{
			name: "Invalid credentials - user not found",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().GetUserByEmail(ctx, "test@example.com").
					Return(nil, sql.ErrNoRows)
			},
			expectedResp: nil,
			expectedErr:  errs.InvalidCredentials,
		},
		{
			name: "Invalid credentials - wrong password",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				hashedPassword, _ := security.HashPassword("password123")
				mockAuthRepo.EXPECT().GetUserByEmail(ctx, "test@example.com").
					Return(&domain.User{
						ID:           "user123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)
			},
			expectedResp: nil,
			expectedErr:  errs.InvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.Login(ctx, tt.req)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.NotEmpty(t, resp.Token)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuth(ctrl)
	mockBudgetRepo := mock2.NewMockBudget(ctrl)
	service := NewService(mockAuthRepo, mockBudgetRepo)
	ctx := context.Background()

	tests := []struct {
		name         string
		req          *LogoutRequest
		mockSetup    func()
		expectedResp *LogoutResponse
		expectedErr  error
	}{
		{
			name: "Successful logout",
			req: &LogoutRequest{
				AuthToken: "valid_token",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().IsTokenBlacklisted(ctx, "valid_token").
					Return(false, nil)
				mockAuthRepo.EXPECT().AddTokenToBlacklist(ctx, "valid_token", gomock.Any()).
					Return(nil)
			},
			expectedResp: &LogoutResponse{Message: "Successfully logged out"},
			expectedErr:  nil,
		},
		{
			name: "Token already blacklisted",
			req: &LogoutRequest{
				AuthToken: "blacklisted_token",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().IsTokenBlacklisted(ctx, "blacklisted_token").
					Return(true, nil)
			},
			expectedResp: &LogoutResponse{Message: "Token is already blacklisted"},
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.Logout(ctx, tt.req)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuth(ctrl)
	mockBudgetRepo := mock2.NewMockBudget(ctrl)
	service := NewService(mockAuthRepo, mockBudgetRepo)
	ctx := context.Background()

	validToken, err := security.GenerateJWT("user123", "test@example.com")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name         string
		req          *RefreshTokenRequest
		mockSetup    func()
		expectedResp *RefreshTokenResponse
		expectedErr  error
	}{
		{
			name: "Successful token refresh",
			req: &RefreshTokenRequest{
				AuthToken: validToken,
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().IsTokenBlacklisted(ctx, validToken).
					Return(false, nil)
				mockAuthRepo.EXPECT().RemoveToken(ctx, validToken).
					Return(nil)
			},
			expectedResp: &RefreshTokenResponse{Token: "mocked_jwt_token"},
			expectedErr:  nil,
		},
		{
			name: "Blacklisted token",
			req: &RefreshTokenRequest{
				AuthToken: "blacklisted_token",
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().IsTokenBlacklisted(ctx, "blacklisted_token").
					Return(true, nil)
			},
			expectedResp: nil,
			expectedErr:  errs.TokenBlacklisted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.RefreshToken(ctx, tt.req)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)

				claims, err := security.GetUserFromToken(resp.Token)
				assert.NoError(t, err)
				assert.Equal(t, "user123", claims.UserID)
				assert.Equal(t, "test@example.com", claims.Email)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuth(ctrl)
	mockBudgetRepo := mock2.NewMockBudget(ctrl)
	service := NewService(mockAuthRepo, mockBudgetRepo)
	ctx := context.Background()

	fakeToken, err := security.GenerateJWT("user123", "test@example.com")
	require.NoError(t, err)

	tests := []struct {
		name         string
		req          *MeRequest
		mockSetup    func()
		expectedResp *MeResponse
		expectedErr  error
	}{
		{
			name: "Successful user info retrieval",
			req: &MeRequest{
				AuthToken: fakeToken,
			},
			mockSetup: func() {
				mockAuthRepo.EXPECT().GetUserByID(ctx, "user123").
					Return(&domain.User{
						ID:        "user123",
						Email:     "test@example.com",
						FirstName: "John",
						LastName:  "Doe",
					}, nil)
			},
			expectedResp: &MeResponse{
				UserInfo: UserInfo{
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.Me(ctx, tt.req)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}
