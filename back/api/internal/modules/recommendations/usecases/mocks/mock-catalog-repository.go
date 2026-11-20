package mocks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockCatalogRepository struct {
	UpsertWorkFn      func(ctx context.Context, work *models.Work) (bool, error)
	GetWorksFn        func(ctx context.Context, keys []string) (map[string]models.Work, error)
	ListUnembeddedFn  func(ctx context.Context, limit int) ([]models.Work, error)
	SaveEmbeddingFn   func(ctx context.Context, key string, embedding models.Embedding) error
	ListVectorsFn     func(ctx context.Context) ([]models.WorkVector, error)
	PruneCandidatesFn func(ctx context.Context, before time.Time) (int, error)
	TopSubjectsFn     func(ctx context.Context, limit int) ([]string, error)
	TopAuthorKeysFn   func(ctx context.Context, limit int) ([]string, error)
}

func (m *MockCatalogRepository) UpsertWork(ctx context.Context, work *models.Work) (bool, error) {
	if m.UpsertWorkFn != nil {
		return m.UpsertWorkFn(ctx, work)
	}
	return true, nil
}

func (m *MockCatalogRepository) GetWorks(ctx context.Context, keys []string) (map[string]models.Work, error) {
	if m.GetWorksFn != nil {
		return m.GetWorksFn(ctx, keys)
	}
	return map[string]models.Work{}, nil
}

func (m *MockCatalogRepository) ListUnembedded(ctx context.Context, limit int) ([]models.Work, error) {
	if m.ListUnembeddedFn != nil {
		return m.ListUnembeddedFn(ctx, limit)
	}
	return nil, nil
}

func (m *MockCatalogRepository) SaveEmbedding(ctx context.Context, key string, embedding models.Embedding) error {
	if m.SaveEmbeddingFn != nil {
		return m.SaveEmbeddingFn(ctx, key, embedding)
	}
	return nil
}

func (m *MockCatalogRepository) ListVectors(ctx context.Context) ([]models.WorkVector, error) {
	if m.ListVectorsFn != nil {
		return m.ListVectorsFn(ctx)
	}
	return nil, nil
}

func (m *MockCatalogRepository) PruneCandidates(ctx context.Context, before time.Time) (int, error) {
	if m.PruneCandidatesFn != nil {
		return m.PruneCandidatesFn(ctx, before)
	}
	return 0, nil
}

func (m *MockCatalogRepository) TopSubjects(ctx context.Context, limit int) ([]string, error) {
	if m.TopSubjectsFn != nil {
		return m.TopSubjectsFn(ctx, limit)
	}
	return nil, nil
}

func (m *MockCatalogRepository) TopAuthorKeys(ctx context.Context, limit int) ([]string, error) {
	if m.TopAuthorKeysFn != nil {
		return m.TopAuthorKeysFn(ctx, limit)
	}
	return nil, nil
}
