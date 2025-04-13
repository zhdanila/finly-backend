package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func RecoverMiddleware() echo.MiddlewareFunc {
	config := middleware.DefaultRecoverConfig
	config.LogErrorFunc = func(c echo.Context, err error, stack []byte) error {
		zap.L().Sugar().Errorf("PANIC RECOVER: %v %s", err, stack)
		return err
	}

	return middleware.RecoverWithConfig(config)
}

func CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	})
}
