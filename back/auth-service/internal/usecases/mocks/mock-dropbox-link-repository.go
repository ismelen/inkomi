package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockDropboxLinkRepository struct {
	GetFn          func(ctx context.Context, userId int) (*models.DropboxLink, error)
	SaveFn         func(ctx context.Context, link *models.DropboxLink) error
	UpdateAccessFn func(ctx context.Context, userId int, access models.AccessToken) error
	UpdateFolderFn func(ctx context.Context, userId int, folder string) error
	DeleteFn       func(ctx context.Context, userId int) error
}

func (m *MockDropboxLinkRepository) Get(ctx context.Context, userId int) (*models.DropboxLink, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, userId)
	}
	return nil, models.ErrDropboxNotLinked
}

func (m *MockDropboxLinkRepository) Save(ctx context.Context, link *models.DropboxLink) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, link)
	}
	return nil
}

func (m *MockDropboxLinkRepository) UpdateAccess(ctx context.Context, userId int, access models.AccessToken) error {
	if m.UpdateAccessFn != nil {
		return m.UpdateAccessFn(ctx, userId, access)
	}
	return nil
}

func (m *MockDropboxLinkRepository) UpdateFolder(ctx context.Context, userId int, folder string) error {
	if m.UpdateFolderFn != nil {
		return m.UpdateFolderFn(ctx, userId, folder)
	}
	return nil
}

func (m *MockDropboxLinkRepository) Delete(ctx context.Context, userId int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userId)
	}
	return nil
}
