package transaction

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	NotEnoughBudget *echo.HTTPError
}{
	NotEnoughBudget: echo.NewHTTPError(http.StatusBadRequest, "Not enough budget"),
}
