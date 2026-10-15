package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
)

type MockUserSettings struct {
	GetDropboxAccountFn func(ctx context.Context, userId int) (*models.DropboxAccount, error)
	GetFCMTokenFn       func(ctx context.Context, userId int) (string, error)
}

func (m *MockUserSettings) GetDropboxAccount(ctx context.Context, userId int) (*models.DropboxAccount, error) {
	if m.GetDropboxAccountFn != nil {
		return m.GetDropboxAccountFn(ctx, userId)
	}
	return nil, models.ErrDropboxNotLinked
}

func (m *MockUserSettings) GetFCMToken(ctx context.Context, userId int) (string, error) {
	if m.GetFCMTokenFn != nil {
		return m.GetFCMTokenFn(ctx, userId)
	}
	return "", models.ErrFCMTokenNotFound
}
