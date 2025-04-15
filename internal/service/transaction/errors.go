package transaction

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	NotEnoughBudget        *echo.HTTPError
	InsufficientBalance    *echo.HTTPError
	InvalidTransactionType *echo.HTTPError
	InvalidInput           *echo.HTTPError
	DatabaseError          *echo.HTTPError
}{
	NotEnoughBudget:        echo.NewHTTPError(http.StatusBadRequest, "Not enough budget"),
	InsufficientBalance:    echo.NewHTTPError(http.StatusBadRequest, "Insufficient balance"),
	InvalidTransactionType: echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction type"),
	InvalidInput:           echo.NewHTTPError(http.StatusBadRequest, "Invalid input provided"),
	DatabaseError:          echo.NewHTTPError(http.StatusInternalServerError, "Database operation failed"),
}
