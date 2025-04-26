package auth

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/repository"
	"finly-backend/pkg/security"
	"go.uber.org/zap"
	"strings"
)

type Service struct {
	authRepo   repository.Auth
	budgetRepo repository.Budget
}

func NewService(authRepo repository.Auth, budgetRepo repository.Budget) *Service {
	return &Service{
		authRepo:   authRepo,
		budgetRepo: budgetRepo,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var err error

	zap.L().Sugar().Infof("Starting user registration for email: %s", req.Email)

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		zap.L().Sugar().Errorf("Error hashing password for email: %s, error: %v", req.Email, err)
		return nil, err
	}

	userID, err := s.authRepo.Register(ctx, req.Email, hashedPassword, req.FirstName, req.LastName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			zap.L().Sugar().Warnf("User already exists with email: %s", req.Email)
			return nil, errs.UserAlreadyExists
		}
		zap.L().Sugar().Errorf("Error registering user with email: %s, error: %v", req.Email, err)
		return nil, err
	}

	token, err := security.GenerateJWT(userID, req.Email)
	if err != nil {
		zap.L().Sugar().Errorf("Error generating JWT for userID: %s, email: %s, error: %v", userID, req.Email, err)
		return nil, err
	}

	zap.L().Sugar().Infof("User registered successfully for email: %s", req.Email)
	return &RegisterResponse{Token: token}, nil
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			zap.L().Sugar().Warnf("Invalid credentials for email: %s", req.Email)
			return nil, errs.InvalidCredentials
		}
		zap.L().Sugar().Errorf("Error fetching user by email: %s, error: %v", req.Email, err)
		return nil, err
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		zap.L().Sugar().Warnf("Invalid password attempt for email: %s", req.Email)
		return nil, errs.InvalidCredentials
	}

	token, err := security.GenerateJWT(user.ID, user.Email)
	if err != nil {
		zap.L().Sugar().Errorf("Error generating JWT for userID: %s, email: %s, error: %v", user.ID, user.Email, err)
		return nil, err
	}

	zap.L().Sugar().Infof("User logged in successfully for email: %s", req.Email)
	return &LoginResponse{Token: token}, nil
}

func (s *Service) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	isBlacklisted, err := s.authRepo.IsTokenBlacklisted(ctx, req.AuthToken)
	if err != nil {
		zap.L().Sugar().Errorf("Error checking if token is blacklisted, token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	if isBlacklisted {
		zap.L().Sugar().Infof("Token already blacklisted, token: %s", req.AuthToken)
		return &LogoutResponse{Message: "Token is already blacklisted"}, nil
	}

	err = s.authRepo.AddTokenToBlacklist(ctx, req.AuthToken, security.TokenTTL.Seconds())
	if err != nil {
		zap.L().Sugar().Errorf("Error blacklisting token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Token blacklisted successfully, token: %s", req.AuthToken)
	return &LogoutResponse{Message: "Successfully logged out"}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	isBlacklisted, err := s.authRepo.IsTokenBlacklisted(ctx, req.AuthToken)
	if err != nil {
		zap.L().Sugar().Errorf("Error checking if token is blacklisted, token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	if isBlacklisted {
		zap.L().Sugar().Warnf("Token is blacklisted, token: %s", req.AuthToken)
		return nil, errs.TokenBlacklisted
	}

	user, err := security.GetUserFromToken(req.AuthToken)
	if err != nil {
		zap.L().Sugar().Errorf("Error extracting user from token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	newToken, err := security.GenerateJWT(user.ID, user.Email)
	if err != nil {
		zap.L().Sugar().Errorf("Error generating new JWT for userID: %s, email: %s, error: %v", user.ID, user.Email, err)
		return nil, err
	}

	err = s.authRepo.RemoveToken(ctx, req.AuthToken)
	if err != nil {
		zap.L().Sugar().Errorf("Error removing old token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Token refreshed successfully for userID: %s", user.ID)
	return &RefreshTokenResponse{Token: newToken}, nil
}

func (s *Service) Me(ctx context.Context, req *MeRequest) (*MeResponse, error) {
	user, err := security.GetUserFromToken(req.AuthToken)
	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			zap.L().Sugar().Warnf("Token expired, token: %s", req.AuthToken)
			return nil, errs.TokenExpired
		}
		zap.L().Sugar().Errorf("Error extracting user from token: %s, error: %v", req.AuthToken, err)
		return nil, err
	}

	userInfo, err := s.authRepo.GetUserByID(ctx, user.UserID)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			zap.L().Sugar().Warnf("User not found for userID: %s", user.UserID)
			return nil, errs.UserNotFound
		}
		zap.L().Sugar().Errorf("Error fetching user info for userID: %s, error: %v", user.UserID, err)
		return nil, err
	}

	zap.L().Sugar().Infof("Successfully retrieved user info for userID: %s", user.UserID)
	return &MeResponse{
		UserInfo: UserInfo{
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			Email:     userInfo.Email,
		},
	}, nil
}
