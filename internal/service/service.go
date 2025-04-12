package service

import (
	"finly-backend/internal/repository"
	"finly-backend/internal/service/auth"
)

type Service struct {
	Auth *auth.Service
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Auth: auth.NewService(repos.Auth, repos.Token),
	}
}
