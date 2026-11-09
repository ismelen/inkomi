package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/domain/models"
)

type ISampleSource interface {
	Get(ctx context.Context, sample models.PreviewSample) ([]byte, error)
}
