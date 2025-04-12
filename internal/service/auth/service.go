package auth

import (
	"context"
	"finly-backend/internal/repository"
	"finly-backend/pkg/security"
	"go.uber.org/zap"
)

type Service struct {
	AuthRepo repository.Auth
}

func NewService(repo repository.Auth) *Service {
	return &Service{
		AuthRepo: repo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var err error

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		zap.L().Error("error hashing password", zap.Error(err))
		return nil, err
	}

	err = s.AuthRepo.Register(ctx, req.Email, hashedPassword, req.FirstName, req.LastName)
	if err != nil {
		zap.L().Error("error registering user", zap.Error(err))
		return nil, err
	}

	return &RegisterResponse{}, nil
}
