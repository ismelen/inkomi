package getsample

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

type GetSampleUC struct {
	sampleSource ports.ISampleSource
}

func NewGetSampleUC(sampleSource ports.ISampleSource) *GetSampleUC {
	return &GetSampleUC{sampleSource: sampleSource}
}

// Execute returns the original image of the preview sample.
func (g *GetSampleUC) Execute(ctx context.Context, sample models.PreviewSample) ([]byte, error) {
	return g.sampleSource.Get(ctx, sample)
}
