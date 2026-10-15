package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

// MockPageProcessor defaults to returning the input as a single JPEG page.
type MockPageProcessor struct {
	ProcessFn       func(ctx context.Context, data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error)
	ProcessSingleFn func(ctx context.Context, data []byte, cfg models.MangaConfig) (models.ProcessedPage, error)
}

func (m *MockPageProcessor) Process(ctx context.Context, data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error) {
	if m.ProcessFn != nil {
		return m.ProcessFn(ctx, data, cfg, rtl)
	}
	return []models.ProcessedPage{{Data: data, Width: 1, Height: 1, Format: models.ImageFormatJPEG}}, nil
}

func (m *MockPageProcessor) ProcessSingle(ctx context.Context, data []byte, cfg models.MangaConfig) (models.ProcessedPage, error) {
	if m.ProcessSingleFn != nil {
		return m.ProcessSingleFn(ctx, data, cfg)
	}
	return models.ProcessedPage{Data: data, Width: 1, Height: 1, Format: models.ImageFormatJPEG}, nil
}
