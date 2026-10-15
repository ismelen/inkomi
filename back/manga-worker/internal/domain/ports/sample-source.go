package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

type ISampleSource interface {
	Get(ctx context.Context, sample models.PreviewSample) ([]byte, error)
}
