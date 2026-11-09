package ports

import "github.com/ismelen/inkomi/back/monolith/internal/modules/epub/domain/models"

type IEpubMerger interface {
	Merge(outputPath string, meta models.EpubMeta, books []models.EpubBook) error
}
