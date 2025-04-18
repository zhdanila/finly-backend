package middleware

import (
	jwt "finly-backend/pkg/security"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const headerUserId = "User-Id"

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
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://frontend.finly.click"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	})
}

func JWT() func(next echo.HandlerFunc) echo.HandlerFunc {
	return echojwt.WithConfig(echojwt.Config{
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			claims, err := jwt.Verify(auth)
			if err != nil {
				return nil, err
			}
			c.Request().Header.Set(headerUserId, claims.UserID)
			return claims, nil
		},
	})
}
