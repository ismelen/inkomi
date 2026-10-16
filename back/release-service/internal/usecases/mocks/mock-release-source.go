package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
)

type MockReleaseSource struct {
	LatestFn func(ctx context.Context) (*models.Release, error)
}

func (m *MockReleaseSource) Latest(ctx context.Context) (*models.Release, error) {
	if m.LatestFn != nil {
		return m.LatestFn(ctx)
	}
	return nil, models.ErrReleaseNotFound
}
