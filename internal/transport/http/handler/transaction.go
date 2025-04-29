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
	group.GET("", s.List)
	group.PATCH("/:id", s.Update)
	group.DELETE("/:id", s.Delete)
}

// @Summary Create a new transaction
// @Description Creates a new transaction for the user with the provided details
// @Tags Transaction
// @ID create-transaction
// @Produce json
// @Param transaction body transaction.CreateTransactionRequest true "TransactionObject Details"
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

	return c.JSON(http.StatusCreated, res)
}

// @Summary List transactions
// @Description Retrieves a list of transactions for the user
// @Tags Transaction
// @ID list-transactions
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} transaction.ListTransactionResponse
// @Router /transaction [get]
func (s *Transaction) List(c echo.Context) error {
	var (
		err error
		obj transaction.ListTransactionRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Transaction.List(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error list budget", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Update a transaction
// @Description Updates an existing transaction with the provided details
// @Tags Transaction
// @ID update-transaction
// @Produce json
// @Param id path string true "TransactionObject ID"
// @Param transaction body transaction.UpdateTransactionRequest true "TransactionObject Details"
// @Success 200 {object} transaction.UpdateTransactionResponse
// @Router /transaction/{id} [patch]
func (s *Transaction) Update(c echo.Context) error {
	var (
		err error
		obj transaction.UpdateTransactionRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Transaction.Update(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error list budget", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Delete a transaction
// @Description Deletes an existing transaction by its ID
// @Tags Transaction
// @ID delete-transaction
// @Produce json
// @Param id path string true "TransactionObject ID"
// @Success 200 {object} transaction.DeleteTransactionResponse
// @Router /transaction/{id} [delete]
func (s *Transaction) Delete(c echo.Context) error {
	var (
		err error
		obj transaction.DeleteTransactionRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Transaction.Delete(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error list budget", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
