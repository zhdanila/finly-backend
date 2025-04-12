package bind

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

type BindOption func(c echo.Context, obj any) error

func BindValidate(c echo.Context, obj any, opts ...BindOption) error {
	var err error

	err = c.Bind(obj)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").SetInternal(err)
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			if err = opt(c, obj); err != nil {
				return err
			}
		}
	}

	err = c.Validate(obj)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return echo.NewHTTPError(http.StatusBadRequest, validationErrors)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	return nil
}

func FromHeaders() func(c echo.Context, obj any) error {
	return func(c echo.Context, obj any) error {
		return (&echo.DefaultBinder{}).BindHeaders(c, obj)
	}
}

func FromQuery() func(c echo.Context, obj any) error {
	return func(c echo.Context, obj any) error {
		return (&echo.DefaultBinder{}).BindQueryParams(c, &obj)
	}
}
