package mocks

import (
	"context"
	"io"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
)

type MockCoverSource struct {
	CoverFn func(ctx context.Context, url string) (*models.Cover, error)
}

func (m *MockCoverSource) Cover(ctx context.Context, url string) (*models.Cover, error) {
	if m.CoverFn != nil {
		return m.CoverFn(ctx, url)
	}
	return &models.Cover{Body: io.NopCloser(strings.NewReader("jpg")), ContentType: "image/jpeg"}, nil
}
