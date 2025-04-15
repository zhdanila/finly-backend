package handler

import (
	"finly-backend/internal/service"
	"finly-backend/internal/service/category"
	"finly-backend/internal/transport/http/middleware"
	"finly-backend/pkg/bind"
	"finly-backend/pkg/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type Category struct {
	service *service.Service
}

func NewCategory(s *service.Service) *Category {
	return &Category{
		service: s,
	}
}

func (s *Category) Register(server *server.Server) {
	group := server.Group("/category", middleware.JWT())

	group.POST("", s.Create)
	group.GET("/:id", s.GetByID)
	group.GET("", s.List)
	group.GET("/custom", s.ListCustom)
	group.DELETE("/:id", s.Delete)
}

// @Summary Create a new category
// @Description Creates a new category for the user with the provided details
// @Tags Category
// @ID create-category
// @Produce json
// @Param category body category.CreateCategoryRequest true "Category Details"
// @Success 200 {object} category.CreateCategoryResponse
// @Router /category [post]
func (s *Category) Create(c echo.Context) error {
	var (
		err error
		obj category.CreateCategoryRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Category.Create(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error creating category", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Get category by ID
// @Description Retrieves the category details for the given ID
// @Tags Category
// @ID get-category-by-id
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} category.GetCategoryByIDResponse
// @Router /category/{id} [get]
func (s *Category) GetByID(c echo.Context) error {
	var (
		err error
		obj category.GetCategoryByIDRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders(), bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Category.GetByID(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error getting category by id", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary List all categories
// @Description Retrieves all categories for the user
// @Tags Category
// @ID list-categories
// @Produce json
// @Success 200 {object} []category.GetCategoryByIDResponse
// @Router /category [get]
func (s *Category) List(c echo.Context) error {
	var (
		err error
		obj category.ListCategoriesRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders(), bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Category.List(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error list categories by id", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary Delete a category
// @Description Deletes the category with the given ID
// @Tags Category
// @ID delete-category
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} category.DeleteCategoryResponse
// @Router /category/{id} [delete]
func (s *Category) Delete(c echo.Context) error {
	var (
		err error
		obj category.DeleteCategoryRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders(), bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Category.Delete(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error deleting category by id", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// @Summary List custom categories
// @Description Retrieves custom categories for the user
// @Tags Category
// @ID list-custom-categories
// @Produce json
// @Success 200 {object} []category.ListCustomCategoriesResponse
func (s *Category) ListCustom(c echo.Context) error {
	var (
		err error
		obj category.ListCustomCategoriesRequest
	)

	if err = bind.Validate(c, &obj, bind.FromHeaders(), bind.FromHeaders()); err != nil {
		zap.L().Error("error binding and validating request", zap.Error(err))
		return err
	}

	res, err := s.service.Category.ListCustom(c.Request().Context(), &obj)
	if err != nil {
		zap.L().Error("error list custom categories", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, res)
}
