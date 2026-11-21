package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type MockCatalogLookup struct {
	BooksFn   func(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	AuthorsFn func(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	WorksFn   func(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error)
}

func (m *MockCatalogLookup) Books(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
	if m.BooksFn != nil {
		return m.BooksFn(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockCatalogLookup) Authors(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
	if m.AuthorsFn != nil {
		return m.AuthorsFn(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockCatalogLookup) AuthorWorks(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
	if m.WorksFn != nil {
		return m.WorksFn(ctx, query, limit)
	}
	return models.WorksPage{Page: query.Page}, nil
}
