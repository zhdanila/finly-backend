package handler

import (
	"finly-backend/internal/service"
	"finly-backend/internal/service/transaction"
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

func NewTransaction(s *service.Service) *Transaction {
	return &Transaction{
		service: s,
	}
}

func (s *Transaction) Register(server *server.Server) {
	group := server.Group("/transaction", middleware.JWT())

	group.POST("", s.Create)
}

// @Summary Create a new transaction
// @Description Creates a new transaction for the user with the provided details
// @Tags Transaction
// @ID create-transaction
// @Produce json
// @Param transaction body transaction.CreateTransactionRequest true "Transaction Details"
// @Success 200 {object} transaction.CreateTransactionResponse
// @Router /transaction [post]
func (s *Transaction) Create(c echo.Context) error {
	var (
		err error
		obj transaction.CreateTransactionRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Transaction.Create(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error creating budget", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
