package handler

import (
	"finly-backend/internal/service"
	"finly-backend/internal/service/budget"
	"finly-backend/internal/transport/http/middleware"
	"finly-backend/pkg/bind"
	"finly-backend/pkg/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type Budget struct {
	service *service.Service
}

func NewBudget(s *service.Service) *Budget {
	return &Budget{
		service: s,
	}
}

func (s *Budget) Register(server *server.Server) {
	group := server.Group("/budget", middleware.JWT())

	group.POST("", s.Create)
}

// @Summary RegisterUser a new user
// @Description Registers a new user with the provided details
// @Tags User
// @ID register-user
// @Produce json
// @Param user body auth.RegisterRequest true "User Details"
// @Success 201 {object} auth.RegisterResponse
// @Router /auth/register [post]
func (s *Budget) Create(c echo.Context) error {
	var (
		err error
		obj budget.CreateBudgetRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Budget.Create(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error registering user", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
