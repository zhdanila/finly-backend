package handler

import (
	"finly-backend/internal/service/auth"
	"finly-backend/pkg/bind"
	"finly-backend/pkg/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type Auth struct {
	service *auth.Service
}

func NewAuth(s *auth.Service) *Auth {
	return &Auth{
		service: s,
	}
}

func (s *Auth) Register(server *server.Server) {
	group := server.Group("/auth")

	group.POST("/register", s.RegisterUser)
}

// @Summary RegisterUser a new user
// @Description Registers a new user with the provided details
// @Tags User
// @ID register-user
// @Produce json
// @Param user body auth.RegisterUserRequest true "User Details"
// @Success 201 {object} auth.RegisterUserResponse
// @Router /auth/register [post]
func (s *Auth) RegisterUser(c echo.Context) error {
	var (
		err error
		obj auth.RegisterRequest
	)

	if err = bind.BindValidate(c, &obj); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.RegisterUser(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error registering user", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
