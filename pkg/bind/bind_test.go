package bind

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"email"`
	Age   int    `json:"age" header:"Age" query:"age"`
}

func setupEchoContext(method, path, body string, headers map[string]string) (*echo.Echo, echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = &structValidator{validator: validator.New()}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return e, c, rec
}

type structValidator struct {
	validator *validator.Validate
}

func (sv *structValidator) Validate(i interface{}) error {
	return sv.validator.Struct(i)
}

func TestValidate_Success(t *testing.T) {
	_, c, _ := setupEchoContext(http.MethodPost, "/test", `{"name":"John","email":"john@example.com"}`, map[string]string{"Age": "30"})
	query := c.Request().URL.Query()
	query.Set("age", "30")
	c.Request().URL.RawQuery = query.Encode()

	obj := &testStruct{}
	err := Validate(c, obj, FromHeaders(), FromQuery())

	assert.NoError(t, err)
	assert.Equal(t, "John", obj.Name)
	assert.Equal(t, "john@example.com", obj.Email)
	assert.Equal(t, 30, obj.Age)
}

func TestValidate_BindError(t *testing.T) {
	_, c, _ := setupEchoContext(http.MethodPost, "/test", `{"name":invalid}`, nil)
	obj := &testStruct{}
	err := Validate(c, obj)

	var httpErr *echo.HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, "Internal server error", httpErr.Message)
}

func TestValidate_ValidationError(t *testing.T) {
	_, c, _ := setupEchoContext(http.MethodPost, "/test", `{"name":"","email":"invalid"}`, nil)
	obj := &testStruct{}
	err := Validate(c, obj)

	var httpErr *echo.HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.IsType(t, validator.ValidationErrors{}, httpErr.Message)
}

func TestValidate_OptionError(t *testing.T) {
	_, c, _ := setupEchoContext(http.MethodPost, "/test", `{"name":"John","email":"john@example.com"}`, nil)
	obj := &testStruct{}
	mockOption := func(c echo.Context, obj any) error {
		return echo.NewHTTPError(http.StatusBadRequest, "Option error")
	}
	err := Validate(c, obj, mockOption)

	var httpErr *echo.HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, "Option error", httpErr.Message)
}

func TestFromHeaders_Success(t *testing.T) {
	_, c, _ := setupEchoContext(http.MethodPost, "/test", `{"name":"John","email":"john@example.com"}`, map[string]string{"Age": "25"})
	obj := &testStruct{}
	err := FromHeaders()(c, obj)

	assert.NoError(t, err)
	assert.Equal(t, 25, obj.Age)
}
