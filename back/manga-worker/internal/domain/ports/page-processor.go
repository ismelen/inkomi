package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

type IPageProcessor interface {
	// Process returns the page ready for the e-reader, or both halves in reading order when
	// it's a split spread.
	Process(ctx context.Context, data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error)
	// ProcessSingle processes data as one page, never splitting or rotating it.
	ProcessSingle(ctx context.Context, data []byte, cfg models.MangaConfig) (models.ProcessedPage, error)
}
