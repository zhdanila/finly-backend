package category

import (
	"context"
	"database/sql"
	"errors"
	"finly-backend/internal/domain"
	"finly-backend/internal/repository/category/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockCategoryRepo := mock.NewMockCategory(ctrl)

	tests := []struct {
		name        string
		req         *CreateCategoryRequest
		mockSetup   func()
		expectedRes *CreateCategoryResponse
		expectedErr error
	}{
		{
			name: "Successful category creation",
			req: &CreateCategoryRequest{
				UserID: "user123",
				CategoryObject: CategoryObject{
					Name: "Groceries",
				},
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().Create(ctx, "user123", "Groceries").
					Return("cat123", nil)
			},
			expectedRes: &CreateCategoryResponse{Id: "cat123"},
			expectedErr: nil,
		},
		{
			name: "Create error",
			req: &CreateCategoryRequest{
				UserID: "user123",
				CategoryObject: CategoryObject{
					Name: "Travel",
				},
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().Create(ctx, "user123", "Travel").
					Return("", errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockCategoryRepo)

			resp, err := service.Create(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockCategoryRepo := mock.NewMockCategory(ctrl)
	createdAt := time.Now()

	tests := []struct {
		name        string
		req         *GetCategoryByIDRequest
		mockSetup   func()
		expectedRes *GetCategoryByIDResponse
		expectedErr error
	}{
		{
			name: "Successful get by ID",
			req: &GetCategoryByIDRequest{
				ID:     "cat123",
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().GetByID(ctx, "cat123", "user123").
					Return(&domain.Category{
						ID:             "cat123",
						UserID:         sql.NullString{String: "user123", Valid: true},
						Name:           "Groceries",
						IsUserCategory: true,
						CreatedAt:      createdAt,
					}, nil)
			},
			expectedRes: &GetCategoryByIDResponse{
				CategoryObject: &CategoryObject{
					ID:             "cat123",
					UserID:         "user123",
					Name:           "Groceries",
					IsUserCategory: true,
					CreatedAt:      createdAt,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Category not found",
			req: &GetCategoryByIDRequest{
				ID:     "cat123",
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().GetByID(ctx, "cat123", "user123").
					Return(nil, sql.ErrNoRows)
			},
			expectedRes: &GetCategoryByIDResponse{},
			expectedErr: nil,
		},
		{
			name: "GetByID error",
			req: &GetCategoryByIDRequest{
				ID:     "cat123",
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().GetByID(ctx, "cat123", "user123").
					Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockCategoryRepo)

			resp, err := service.GetByID(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockCategoryRepo := mock.NewMockCategory(ctrl)
	createdAt := time.Now()

	tests := []struct {
		name        string
		req         *ListCategoriesRequest
		mockSetup   func()
		expectedRes *ListCategoriesResponse
		expectedErr error
	}{
		{
			name: "Successful list categories",
			req: &ListCategoriesRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().List(ctx, "user123").
					Return([]*domain.Category{
						{
							ID:             "cat1",
							UserID:         sql.NullString{String: "user123", Valid: true},
							Name:           "Groceries",
							IsUserCategory: true,
							CreatedAt:      createdAt,
						},
						{
							ID:             "cat2",
							UserID:         sql.NullString{String: "user123", Valid: true},
							Name:           "Bills",
							IsUserCategory: false,
							CreatedAt:      createdAt,
						},
					}, nil)
			},
			expectedRes: &ListCategoriesResponse{
				Categories: []CategoryObject{
					{
						ID:             "cat1",
						UserID:         "user123",
						Name:           "Groceries",
						IsUserCategory: true,
						CreatedAt:      createdAt,
					},
					{
						ID:             "cat2",
						UserID:         "user123",
						Name:           "Bills",
						IsUserCategory: false,
						CreatedAt:      createdAt,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "List error",
			req: &ListCategoriesRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().List(ctx, "user123").
					Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockCategoryRepo)

			resp, err := service.List(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestListCustom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockCategoryRepo := mock.NewMockCategory(ctrl)
	createdAt := time.Now()

	tests := []struct {
		name        string
		req         *ListCustomCategoriesRequest
		mockSetup   func()
		expectedRes *ListCustomCategoriesResponse
		expectedErr error
	}{
		{
			name: "Successful list custom categories",
			req: &ListCustomCategoriesRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().ListCustom(ctx, "user123").
					Return([]*domain.Category{
						{
							ID:             "cat1",
							UserID:         sql.NullString{String: "user123", Valid: true},
							Name:           "Gadgets",
							IsUserCategory: true,
							CreatedAt:      createdAt,
						},
					}, nil)
			},
			expectedRes: &ListCustomCategoriesResponse{
				Categories: []CategoryObject{
					{
						ID:             "cat1",
						UserID:         "user123",
						Name:           "Gadgets",
						IsUserCategory: true,
						CreatedAt:      createdAt,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "ListCustom error",
			req: &ListCustomCategoriesRequest{
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().ListCustom(ctx, "user123").
					Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockCategoryRepo)

			resp, err := service.ListCustom(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockCategoryRepo := mock.NewMockCategory(ctrl)

	tests := []struct {
		name        string
		req         *DeleteCategoryRequest
		mockSetup   func()
		expectedRes *DeleteCategoryResponse
		expectedErr error
	}{
		{
			name: "Successful delete",
			req: &DeleteCategoryRequest{
				ID:     "cat123",
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().Delete(ctx, "cat123", "user123").
					Return(nil)
			},
			expectedRes: &DeleteCategoryResponse{},
			expectedErr: nil,
		},
		{
			name: "Delete error",
			req: &DeleteCategoryRequest{
				ID:     "cat123",
				UserID: "user123",
			},
			mockSetup: func() {
				mockCategoryRepo.EXPECT().Delete(ctx, "cat123", "user123").
					Return(errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			service := NewService(mockCategoryRepo)

			resp, err := service.Delete(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}
