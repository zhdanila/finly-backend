package router

import (
	_ "finly-backend/docs"
	"finly-backend/internal/service"
	"finly-backend/internal/transport/http/handler"
	"finly-backend/pkg/server"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"net/http"
)

func RegisterRoutes(server *server.Server, services *service.Service) {
	// Register handlers
	handler.NewAuth(services).Register(server)
	handler.NewCategory(services).Register(server)
	handler.NewBudget(services).Register(server)
	handler.NewTransaction(services).Register(server)

	server.GET("/swagger/*", echoSwagger.WrapHandler)
	server.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
