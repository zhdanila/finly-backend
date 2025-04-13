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
	group.GET("/:budget_id", s.GetByID)
	group.GET("", s.List)
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

// @Summary Get budget by ID
// @Description Retrieves a budget by its ID for the specified user
// @Tags Budget
// @ID get-budget-by-id
// @Produce json
// @Param budget_id path string true "Budget ID"
// @Param user_id header string true "User ID"
// @Success 200 {object} budget.GetBudgetByIDResponse
// @Router /budget/{budget_id} [get]
func (s *Budget) GetByID(c echo.Context) error {
	var (
		err error
		obj budget.GetBudgetByIDRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Budget.GetByID(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error getting budget by id", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary List all budgets
// @Description Retrieves a list of all budgets for the specified user
// @Tags Budget
// @ID list-budgets
// @Produce json
// @Param user_id header string true "User ID"
// @Success 200 {object} budget.ListBudgetsByIDResponse
// @Router /budget [get]
func (s *Budget) List(c echo.Context) error {
	var (
		err error
		obj budget.ListBudgetsByIDRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Budget.List(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error getting budget by id", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
