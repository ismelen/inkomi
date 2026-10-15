package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockUserRepository struct {
	CreateFn         func(ctx context.Context, user *models.User) error
	GetByIdFn        func(ctx context.Context, id int) (*models.User, error)
	GetByEmailFn     func(ctx context.Context, email string) (*models.User, error)
	UpdateFn         func(ctx context.Context, user *models.User) error
	UpdateFCMTokenFn func(ctx context.Context, id int, token string) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	user.Id = 1
	return nil
}

func (m *MockUserRepository) GetById(ctx context.Context, id int) (*models.User, error) {
	if m.GetByIdFn != nil {
		return m.GetByIdFn(ctx, id)
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) UpdateFCMToken(ctx context.Context, id int, token string) error {
	if m.UpdateFCMTokenFn != nil {
		return m.UpdateFCMTokenFn(ctx, id, token)
	}
	return nil
}
