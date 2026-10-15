package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

type MockSampleSource struct {
	GetFn func(ctx context.Context, sample models.PreviewSample) ([]byte, error)
}

func (m *MockSampleSource) Get(ctx context.Context, sample models.PreviewSample) ([]byte, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, sample)
	}
	return []byte(sample), nil
}
