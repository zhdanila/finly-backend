package service

import (
	"finly-backend/internal/repository"
	"finly-backend/internal/service/auth"
	"finly-backend/internal/service/budget"
)

type Service struct {
	Auth   *auth.Service
	Budget *budget.Service
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Auth:   auth.NewService(repos.Auth),
		Budget: budget.NewService(repos.Budget),
	}
}
