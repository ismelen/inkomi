package ports

import "github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"

type EpubMerger interface {
	Merge(outputPath string, meta models.EpubMeta, books []models.EpubBook) error
}
