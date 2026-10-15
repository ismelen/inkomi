package previewpage

import (
	"context"
	"fmt"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

type PreviewPageUC struct {
	sampleSource ports.ISampleSource
	processor    ports.IPageProcessor
}

func NewPreviewPageUC(sampleSource ports.ISampleSource, processor ports.IPageProcessor) *PreviewPageUC {
	return &PreviewPageUC{sampleSource: sampleSource, processor: processor}
}

// Execute processes the preview sample with cfg as a single page.
func (p *PreviewPageUC) Execute(ctx context.Context, sample models.PreviewSample, cfg models.MangaConfig) (models.ProcessedPage, error) {
	if err := cfg.FitsPreview(); err != nil {
		return models.ProcessedPage{}, err
	}

	data, err := p.sampleSource.Get(ctx, sample)
	if err != nil {
		return models.ProcessedPage{}, fmt.Errorf("get sample %s: %w", sample, err)
	}

	page, err := p.processor.ProcessSingle(ctx, data, cfg)
	if err != nil {
		return models.ProcessedPage{}, fmt.Errorf("process sample %s: %w", sample, err)
	}
	return page, nil
}
