package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type MockSourceRepository struct {
	CreateFn                     func(ctx context.Context, source *models.Source) (string, error)
	GetByIdAndUserIdCompactFn    func(ctx context.Context, id string, userID int) (*models.CompactSoruce, error)
	UpdateStatusFn               func(ctx context.Context, id string, userId int, status models.SourceStatus) error
	CheckSiblingsInWaitingJoinFn func(ctx context.Context, id string, userId string) (*models.CompactSoruce, bool, error)
}

func (m *MockSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, source)
	}
	return "", nil
}

func (m *MockSourceRepository) SetError(ctx context.Context, id string, userId int, errMsg string) error {
	return nil
}

func (m *MockSourceRepository) GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error) {
	return nil, nil
}

func (m *MockSourceRepository) GetByIdAndUserIdCompact(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
	if m.GetByIdAndUserIdCompactFn != nil {
		return m.GetByIdAndUserIdCompactFn(ctx, id, userID)
	}
	return nil, nil
}

func (m *MockSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, userId, status)
	}
	return nil
}

func (m *MockSourceRepository) Delete(ctx context.Context, id string, userID int) error {
	return nil
}

func (m *MockSourceRepository) CheckSiblingsInWaitingJoin(ctx context.Context, id string, userId string) (*models.CompactSoruce, bool, error) {
	if m.CheckSiblingsInWaitingJoinFn != nil {
		return m.CheckSiblingsInWaitingJoinFn(ctx, id, userId)
	}
	return nil, false, nil
}
