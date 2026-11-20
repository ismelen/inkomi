package dismissrecommendation

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

// DismissRecommendationUC hides a work from the user's recommendations for good.
type DismissRecommendationUC struct {
	signals ports.ISignalRepository
}

func NewDismissRecommendationUC(signals ports.ISignalRepository) *DismissRecommendationUC {
	return &DismissRecommendationUC{signals: signals}
}

func (d *DismissRecommendationUC) Execute(ctx context.Context, userId int, workKey string) error {
	workKey = strings.TrimSpace(workKey)
	if workKey == "" {
		return models.ErrEmptyWorkKey
	}
	return d.signals.Dismiss(ctx, userId, workKey)
}
