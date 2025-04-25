package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var errs = struct {
	UserAlreadyExists  *echo.HTTPError
	InvalidCredentials *echo.HTTPError
	TokenBlacklisted   *echo.HTTPError
	InvalidToken       *echo.HTTPError
	UserNotFound       *echo.HTTPError
	TokenExpired       *echo.HTTPError
}{
	UserAlreadyExists:  echo.NewHTTPError(http.StatusConflict, "User already exists"),
	InvalidCredentials: echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials"),
	TokenBlacklisted:   echo.NewHTTPError(http.StatusForbidden, "Token is blacklisted"),
	InvalidToken:       echo.NewHTTPError(http.StatusUnauthorized, "Invalid token"),
	UserNotFound:       echo.NewHTTPError(http.StatusNotFound, "User not found"),
	TokenExpired:       echo.NewHTTPError(http.StatusUnauthorized, "Token expired"),
}
