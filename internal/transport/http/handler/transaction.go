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

type Transaction struct {
	service *service.Service
}

func NewTransaction(s *service.Service) *Budget {
	return &Budget{
		service: s,
	}
}

func (s *Transaction) Register(server *server.Server) {
	group := server.Group("/budget", middleware.JWT())

	group.POST("", s.Create)
	group.GET("/:budget_id", s.GetByID)
}

// @Summary Create a new budget
// @Description Creates a new budget for the user with the provided details
// @Tags Budget
// @ID create-budget
// @Produce json
// @Param budget body budget.CreateBudgetRequest true "Budget Details"
// @Success 200 {object} budget.CreateBudgetResponse
// @Router /budget [post]
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
		zap.L().Error("error creating budget", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
