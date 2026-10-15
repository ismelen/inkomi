package mocks

import (
	"os"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

// MockMangaEpubBuilder defaults to writing an empty file at outPath.
type MockMangaEpubBuilder struct {
	BuildFn func(outPath string, meta models.EpubMeta, pages []models.EpubPage) error
}

func (m *MockMangaEpubBuilder) Build(outPath string, meta models.EpubMeta, pages []models.EpubPage) error {
	if m.BuildFn != nil {
		return m.BuildFn(outPath, meta, pages)
	}
	return os.WriteFile(outPath, nil, 0o644)
}
