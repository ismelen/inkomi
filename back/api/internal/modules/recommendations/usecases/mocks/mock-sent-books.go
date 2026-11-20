package mocks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockSentBooks struct {
	ListSinceFn func(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error)
}

func (m *MockSentBooks) ListSince(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error) {
	if m.ListSinceFn != nil {
		return m.ListSinceFn(ctx, since, limit)
	}
	return nil, nil
}
