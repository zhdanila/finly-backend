package router

import (
	"finly-backend/internal/service"
	"finly-backend/internal/transport/http/handler"
	"finly-backend/internal/transport/http/middleware"
	"finly-backend/pkg/server"
)

func RegisterRoutes(server *server.Server, services *service.Service) {
	server.Use(middleware.RecoverMiddleware())

	// Register handlers
	handler.NewAuth(services).Register(server)
}
