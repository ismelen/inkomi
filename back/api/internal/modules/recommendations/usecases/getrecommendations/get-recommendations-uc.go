package getrecommendations

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

type GetRecommendationsUC struct {
	recs ports.IRecommendationRepository
}

func NewGetRecommendationsUC(recs ports.IRecommendationRepository) *GetRecommendationsUC {
	return &GetRecommendationsUC{recs: recs}
}

func (g *GetRecommendationsUC) Execute(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error) {
	if !section.IsValid() {
		return nil, models.ErrInvalidSection
	}
	return g.recs.List(ctx, userId, section)
}
