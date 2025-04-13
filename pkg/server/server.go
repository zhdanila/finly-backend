package server

import (
	"finly-backend/internal/transport/http/middleware"
	validator2 "finly-backend/pkg/validator"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Server struct {
	Port string
	*echo.Echo
}

func NewServer(port string) *Server {
	var err error

	server := echo.New()
	server.Use(middleware.RecoverMiddleware())
	server.Use(middleware.CORSMiddleware())

	if server.Validator, err = validator2.CustomValidator(); err != nil {
		zap.L().Fatal("Error setting up custom validator", zap.Error(err))
	}

	return &Server{
		Port: port,
		Echo: server,
	}
}

func (s *Server) Start() error {
	return s.Echo.Start(fmt.Sprintf(":%s", s.Port))
}
