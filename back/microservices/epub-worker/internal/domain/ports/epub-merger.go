package ports

import "github.com/ismelen/inkomi/back/microservices/epub-worker/internal/domain/models"

type IEpubMerger interface {
	Merge(outputPath string, meta models.EpubMeta, books []models.EpubBook) error
}
