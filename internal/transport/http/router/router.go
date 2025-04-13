package router

import (
	"finly-backend/internal/service"
	"finly-backend/internal/transport/http/handler"
	"finly-backend/pkg/server"
)

func RegisterRoutes(server *server.Server, services *service.Service) {
	// Register handlers
	handler.NewAuth(services).Register(server)
	handler.NewBudget(services).Register(server)
	handler.NewCategory(services).Register(server)
}
