package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	UserAlreadyExists  *echo.HTTPError
	InvalidCredentials *echo.HTTPError
}{
	UserAlreadyExists:  echo.NewHTTPError(http.StatusConflict, "User already exists"),
	InvalidCredentials: echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials"),
}
