package auth

import (
	"context"
	"finly-backend/internal/repository"
	"finly-backend/pkg/security"
	"strings"
)

type Service struct {
	repo  repository.Auth
	token repository.Token
}

func NewService(repo repository.Auth, tokenBlacklist repository.Token) *Service {
	return &Service{
		repo:  repo,
		token: tokenBlacklist,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var err error

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	userID, err := s.repo.Register(ctx, req.Email, hashedPassword, req.FirstName, req.LastName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, errs.UserAlreadyExists
		}

		return nil, err
	}

	token, err := security.GenerateJWT(userID, req.Email)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{Token: token}, nil
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errs.InvalidCredentials
		}
		return nil, err
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errs.InvalidCredentials
	}

	token, err := security.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

func (s *Service) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	isBlacklisted, err := s.token.IsTokenBlacklisted(ctx, req.AuthToken)
	if err != nil {
		return nil, err
	}

	if isBlacklisted {
		return &LogoutResponse{Message: "Token is already blacklisted"}, nil
	}

	err = s.token.AddTokenToBlacklist(ctx, req.AuthToken, security.TokenTTL.Seconds())
	if err != nil {
		return nil, err
	}

	return &LogoutResponse{Message: "Successfully logged out"}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	isBlacklisted, err := s.token.IsTokenBlacklisted(ctx, req.AuthToken)
	if err != nil {
		return nil, err
	}

	if isBlacklisted {
		return nil, errs.TokenBlacklisted
	}

	user, err := security.GetUserFromToken(req.AuthToken)
	if err != nil {
		return nil, err
	}

	newToken, err := security.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	err = s.token.RemoveToken(ctx, req.AuthToken)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{Token: newToken}, nil
}
