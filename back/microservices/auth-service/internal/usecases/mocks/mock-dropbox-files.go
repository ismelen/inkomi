package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"
)

type MockDropboxFiles struct {
	ListFoldersFn  func(ctx context.Context, accessToken, path string) ([]models.DropboxFolder, error)
	CreateFolderFn func(ctx context.Context, accessToken, path string) (*models.DropboxFolder, error)
}

func (m *MockDropboxFiles) ListFolders(ctx context.Context, accessToken, path string) ([]models.DropboxFolder, error) {
	if m.ListFoldersFn != nil {
		return m.ListFoldersFn(ctx, accessToken, path)
	}
	return nil, nil
}

func (m *MockDropboxFiles) CreateFolder(ctx context.Context, accessToken, path string) (*models.DropboxFolder, error) {
	if m.CreateFolderFn != nil {
		return m.CreateFolderFn(ctx, accessToken, path)
	}
	return &models.DropboxFolder{Path: path}, nil
}
