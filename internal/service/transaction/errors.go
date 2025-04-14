package transaction

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	NotEnoughBudget        *echo.HTTPError
	InsufficientBalance    *echo.HTTPError
	InvalidTransactionType *echo.HTTPError
}{
	NotEnoughBudget:        echo.NewHTTPError(http.StatusBadRequest, "Not enough budget"),
	InsufficientBalance:    echo.NewHTTPError(http.StatusBadRequest, "Insufficient balance"),
	InvalidTransactionType: echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction type"),
}
