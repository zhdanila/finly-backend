package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"finly-backend/pkg/security"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRecoverMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(RecoverMiddleware())

	e.GET("/panic", func(c echo.Context) error {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCORSMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "CORS OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set(echo.HeaderOrigin, "http://localhost:5173")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods), http.MethodGet)
}

func TestJWTMiddleware_Success(t *testing.T) {
	e := echo.New()
	e.Use(JWT())

	token, err := security.GenerateJWT("user123", "user@example.com")
	assert.NoError(t, err)

	e.GET("/protected", func(c echo.Context) error {
		userID := c.Request().Header.Get("User-Id")
		return c.String(http.StatusOK, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user123", rec.Body.String())
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	e := echo.New()
	e.Use(JWT())

	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, "should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	e := echo.New()
	e.Use(JWT())

	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, "should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer invalid_token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	e := echo.New()
	e.Use(JWT())

	// Manually create an expired token
	expiredClaims := &security.Claims{
		UserID: "expiredUser",
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, err := tokenObj.SignedString([]byte("your_secret_key"))
	assert.NoError(t, err)

	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, "should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+expiredToken)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
