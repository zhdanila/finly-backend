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
