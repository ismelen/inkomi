package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type MockSourceRepository struct {
	CreateFn                 func(ctx context.Context, source *models.Source) (string, error)
	GetByIDAndUserIDFn       func(ctx context.Context, id string, userID int) (*models.Source, error)
	GetDetailByIDAndUserIDFn func(ctx context.Context, id string, userID int) (*models.SourceDetail, error)
	UpdateStatusFn           func(ctx context.Context, id string, userID int, status models.SourceStatus) error
	DeleteFn                 func(ctx context.Context, id string, userID int) error
}

func (m *MockSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, source)
	}
	return "", nil
}

func (m *MockSourceRepository) GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error) {
	if m.GetByIDAndUserIDFn != nil {
		return m.GetByIDAndUserIDFn(ctx, id, userID)
	}
	return nil, models.ErrSourceNotFound
}

func (m *MockSourceRepository) GetDetailByIDAndUserID(ctx context.Context, id string, userID int) (*models.SourceDetail, error) {
	if m.GetDetailByIDAndUserIDFn != nil {
		return m.GetDetailByIDAndUserIDFn(ctx, id, userID)
	}
	return nil, models.ErrSourceNotFound
}

func (m *MockSourceRepository) UpdateStatus(ctx context.Context, id string, userID int, status models.SourceStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, userID, status)
	}
	return nil
}

func (m *MockSourceRepository) Delete(ctx context.Context, id string, userID int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, userID)
	}
	return nil
}
