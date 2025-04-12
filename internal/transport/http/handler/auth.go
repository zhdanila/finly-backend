package handler

import (
	"finly-backend/internal/service"
	"finly-backend/internal/service/auth"
	"finly-backend/pkg/bind"
	"finly-backend/pkg/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type Auth struct {
	service *service.Service
}

func NewAuth(s *service.Service) *Auth {
	return &Auth{
		service: s,
	}
}

func (s *Auth) Register(server *server.Server) {
	group := server.Group("/auth")

	group.POST("/register", s.RegisterUser)
	group.POST("/login", s.Login)
	group.POST("/logout", s.Logout)
	group.POST("/refresh", s.Refresh)
}

// @Summary RegisterUser a new user
// @Description Registers a new user with the provided details
// @Tags User
// @ID register-user
// @Produce json
// @Param user body auth.RegisterRequest true "User Details"
// @Success 201 {object} auth.RegisterResponse
// @Router /auth/register [post]
func (s *Auth) RegisterUser(c echo.Context) error {
	var (
		err error
		obj auth.RegisterRequest
	)

	if err = bind.Validate(c, &obj); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Auth.Register(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error registering user", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Login a user
// @Description Authenticates a user with the provided credentials
// @Tags User
// @ID login-user
// @Produce json
// @Param credentials body auth.LoginRequest true "User Credentials"
// @Success 200 {object} auth.LoginResponse
// @Router /auth/login [post]
func (s *Auth) Login(c echo.Context) error {
	var (
		err error
		obj auth.LoginRequest
	)

	if err = bind.Validate(c, &obj); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Auth.Login(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error login user", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Logout a user
// @Description Logs out a user by invalidating their authentication token
// @Tags User
// @ID logout-user
// @Produce json
// @Param token body auth.LogoutRequest true "Authentication Token"
// @Success 200 {object} auth.LogoutResponse
// @Router /auth/logout [post]
func (s *Auth) Logout(c echo.Context) error {
	var (
		err error
		obj auth.LogoutRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Auth.Logout(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error logout user", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Refresh a user token
// @Description Refreshes a user's authentication token by generating a new one
// @Tags User
// @ID refresh-token
// @Produce json
// @Param token body auth.RefreshTokenRequest true "Authentication Token"
// @Success 200 {object} auth.RefreshTokenResponse
// @Router /auth/refresh [post]
func (s *Auth) Refresh(c echo.Context) error {
	var (
		err error
		obj auth.RefreshTokenRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Auth.RefreshToken(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error refresh token", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
