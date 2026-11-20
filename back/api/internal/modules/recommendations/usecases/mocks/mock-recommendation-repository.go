package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockRecommendationRepository struct {
	ReplaceFn func(ctx context.Context, userId int, recs []models.Recommendation) error
	ListFn    func(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error)
}

func (m *MockRecommendationRepository) Replace(ctx context.Context, userId int, recs []models.Recommendation) error {
	if m.ReplaceFn != nil {
		return m.ReplaceFn(ctx, userId, recs)
	}
	return nil
}

func (m *MockRecommendationRepository) List(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userId, section)
	}
	return nil, nil
}
