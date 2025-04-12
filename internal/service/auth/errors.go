package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	UserAlreadyExists *echo.HTTPError
}{
	UserAlreadyExists: echo.NewHTTPError(http.StatusNotFound, "User already exists"),
}
