package ports

import "github.com/ismelen/inkomi/back/microservices/manga-worker/internal/domain/models"

type IMangaEpubBuilder interface {
	Build(outPath string, meta models.EpubMeta, pages []models.EpubPage) error
}
