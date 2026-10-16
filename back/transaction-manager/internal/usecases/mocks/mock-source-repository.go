package mocks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type MockSourceRepository struct {
	CreateFn                         func(ctx context.Context, source *models.Source) (string, error)
	GetByIdAndUserIdFn               func(ctx context.Context, id string, userId int) (*models.Source, error)
	GetDetailByIdAndUserIdFn         func(ctx context.Context, id string, userId int) (*models.SourceDetail, error)
	ListTreesByUserIdFn              func(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error)
	NowFn                            func(ctx context.Context) (time.Time, error)
	UpdateStatusFn                   func(ctx context.Context, id string, userId int, status models.SourceStatus) error
	UpdateFileFn                     func(ctx context.Context, id string, userId int, filename string, size int64) error
	UpdateErrorMsgFn                 func(ctx context.Context, id string, userId int, errorMsg string) error
	UpsertResultFn                   func(ctx context.Context, result *models.Source) error
	ListResultsByParentIdAndUserIdFn func(ctx context.Context, parentId string, userId int) ([]models.Source, error)
	DeleteFn                         func(ctx context.Context, id string, userId int) error
}

func (m *MockSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, source)
	}
	return "", nil
}

func (m *MockSourceRepository) GetByIdAndUserId(ctx context.Context, id string, userId int) (*models.Source, error) {
	if m.GetByIdAndUserIdFn != nil {
		return m.GetByIdAndUserIdFn(ctx, id, userId)
	}
	return nil, models.ErrSourceNotFound
}

func (m *MockSourceRepository) GetDetailByIdAndUserId(ctx context.Context, id string, userId int) (*models.SourceDetail, error) {
	if m.GetDetailByIdAndUserIdFn != nil {
		return m.GetDetailByIdAndUserIdFn(ctx, id, userId)
	}
	return nil, models.ErrSourceNotFound
}

func (m *MockSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, userId, status)
	}
	return nil
}

func (m *MockSourceRepository) UpdateFile(ctx context.Context, id string, userId int, filename string, size int64) error {
	if m.UpdateFileFn != nil {
		return m.UpdateFileFn(ctx, id, userId, filename, size)
	}
	return nil
}

func (m *MockSourceRepository) UpdateErrorMsg(ctx context.Context, id string, userId int, errorMsg string) error {
	if m.UpdateErrorMsgFn != nil {
		return m.UpdateErrorMsgFn(ctx, id, userId, errorMsg)
	}
	return nil
}

func (m *MockSourceRepository) UpsertResult(ctx context.Context, result *models.Source) error {
	if m.UpsertResultFn != nil {
		return m.UpsertResultFn(ctx, result)
	}
	return nil
}

func (m *MockSourceRepository) ListResultsByParentIdAndUserId(ctx context.Context, parentId string, userId int) ([]models.Source, error) {
	if m.ListResultsByParentIdAndUserIdFn != nil {
		return m.ListResultsByParentIdAndUserIdFn(ctx, parentId, userId)
	}
	return nil, nil
}

func (m *MockSourceRepository) Delete(ctx context.Context, id string, userId int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, userId)
	}
	return nil
}

func (m *MockSourceRepository) ListTreesByUserId(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
	if m.ListTreesByUserIdFn != nil {
		return m.ListTreesByUserIdFn(ctx, userId, filter)
	}
	return nil, nil
}

func (m *MockSourceRepository) Now(ctx context.Context) (time.Time, error) {
	if m.NowFn != nil {
		return m.NowFn(ctx)
	}
	return time.Now().UTC().Truncate(time.Second), nil
}
