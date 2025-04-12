package auth

import (
	"context"
	"finly-backend/internal/repository"
	"finly-backend/pkg/security"
	"strings"
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
		return nil, err
	}

	userID, err := s.AuthRepo.Register(ctx, req.Email, hashedPassword, req.FirstName, req.LastName)
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

	return &RegisterResponse{
		Token: token,
	}, nil
}
