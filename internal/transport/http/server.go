package http

import (
	"finly-backend/internal/repository/service"
	"finly-backend/internal/transport/http/middleware"
	"fmt"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Port string
	*echo.Echo
}

func NewServer(port string, service *service.Service) *Server {
	server := echo.New()
	server.Use(middleware.RecoverMiddleware())

	// register handlers

	return &Server{
		Port: port,
		Echo: server,
	}
}

func (s *Server) Start() error {
	return s.Echo.Start(fmt.Sprintf(":%s", s.Port))
}
