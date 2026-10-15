package mocks

import (
	"os"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
)

type MockEpubMerger struct {
	MergeFn func(outputPath string, meta models.EpubMeta, books []models.EpubBook) error
}

// Merge writes an empty output file by default so the caller can upload it.
func (m *MockEpubMerger) Merge(outputPath string, meta models.EpubMeta, books []models.EpubBook) error {
	if m.MergeFn != nil {
		return m.MergeFn(outputPath, meta, books)
	}
	return os.WriteFile(outputPath, nil, 0o600)
}
