package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockCatalogSource struct {
	MatchBookFn      func(ctx context.Context, title, author string) (*models.Work, error)
	WorkFn           func(ctx context.Context, key string) (*models.Work, error)
	MatchAuthorFn    func(ctx context.Context, name string) (string, string, error)
	DescriptionFn    func(ctx context.Context, workKey string) (string, error)
	WorksByAuthorFn  func(ctx context.Context, authorKey string, limit int) ([]models.Work, error)
	WorksBySubjectFn func(ctx context.Context, subject string, limit int) ([]models.Work, error)
	NewWorksFn       func(ctx context.Context, subject, authorKey string, sinceYear, limit int) ([]models.Work, error)
}

func (m *MockCatalogSource) MatchBook(ctx context.Context, title, author string) (*models.Work, error) {
	if m.MatchBookFn != nil {
		return m.MatchBookFn(ctx, title, author)
	}
	return nil, nil
}

func (m *MockCatalogSource) Work(ctx context.Context, key string) (*models.Work, error) {
	if m.WorkFn != nil {
		return m.WorkFn(ctx, key)
	}
	return nil, nil
}

func (m *MockCatalogSource) MatchAuthor(ctx context.Context, name string) (string, string, error) {
	if m.MatchAuthorFn != nil {
		return m.MatchAuthorFn(ctx, name)
	}
	return "", "", nil
}

func (m *MockCatalogSource) Description(ctx context.Context, workKey string) (string, error) {
	if m.DescriptionFn != nil {
		return m.DescriptionFn(ctx, workKey)
	}
	return "", nil
}

func (m *MockCatalogSource) WorksByAuthor(ctx context.Context, authorKey string, limit int) ([]models.Work, error) {
	if m.WorksByAuthorFn != nil {
		return m.WorksByAuthorFn(ctx, authorKey, limit)
	}
	return nil, nil
}

func (m *MockCatalogSource) WorksBySubject(ctx context.Context, subject string, limit int) ([]models.Work, error) {
	if m.WorksBySubjectFn != nil {
		return m.WorksBySubjectFn(ctx, subject, limit)
	}
	return nil, nil
}

func (m *MockCatalogSource) NewWorks(ctx context.Context, subject, authorKey string, sinceYear, limit int) ([]models.Work, error) {
	if m.NewWorksFn != nil {
		return m.NewWorksFn(ctx, subject, authorKey, sinceYear, limit)
	}
	return nil, nil
}
