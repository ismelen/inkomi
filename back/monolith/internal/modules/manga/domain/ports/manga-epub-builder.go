package ports

import "github.com/ismelen/inkomi/back/monolith/internal/modules/manga/domain/models"

type IMangaEpubBuilder interface {
	Build(outPath string, meta models.EpubMeta, pages []models.EpubPage) error
}
