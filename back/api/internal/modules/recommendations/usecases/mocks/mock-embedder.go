package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockEmbedder struct {
	EmbedPassagesFn func(ctx context.Context, texts []string) ([]models.Embedding, error)
	Releases        int
}

func (m *MockEmbedder) EmbedPassages(ctx context.Context, texts []string) ([]models.Embedding, error) {
	if m.EmbedPassagesFn != nil {
		return m.EmbedPassagesFn(ctx, texts)
	}
	vectors := make([]models.Embedding, len(texts))
	for i := range texts {
		vectors[i] = models.Embedding{1, 0}
	}
	return vectors, nil
}

func (m *MockEmbedder) Release() {
	m.Releases++
}
