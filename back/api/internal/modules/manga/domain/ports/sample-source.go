package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/manga/domain/models"
)

type ISampleSource interface {
	Get(ctx context.Context, sample models.PreviewSample) ([]byte, error)
}
