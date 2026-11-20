package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type IRecommendationRepository interface {
	// Replace swaps the recommendations of the user for recs (every section at once).
	Replace(ctx context.Context, userId int, recs []models.Recommendation) error
	// List returns the recommendations of a section with their works and reasons, by rank,
	// without the dismissed ones.
	List(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error)
}
