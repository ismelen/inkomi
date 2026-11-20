package embedworks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	batchSize = 16
	// maxPerRun keeps a refresh short; what is left goes in the next one.
	maxPerRun = 1000
)

// EmbedWorksUC computes the embeddings the catalog is missing, user works first.
type EmbedWorksUC struct {
	embedder ports.IEmbedder
	catalog  ports.ICatalogRepository
	// ready holds the model back while the free memory is low (it needs a few hundred MB).
	ready func() bool
}

func NewEmbedWorksUC(embedder ports.IEmbedder, catalog ports.ICatalogRepository, ready func() bool) *EmbedWorksUC {
	return &EmbedWorksUC{embedder: embedder, catalog: catalog, ready: ready}
}

// Execute embeds up to maxPerRun works and returns how many; the model is released afterwards.
func (e *EmbedWorksUC) Execute(ctx context.Context) (int, error) {
	defer e.embedder.Release()

	embedded := 0
	for embedded < maxPerRun {
		if ctx.Err() != nil {
			return embedded, ctx.Err()
		}
		if e.ready != nil && !e.ready() {
			return embedded, nil
		}
		works, err := e.catalog.ListUnembedded(ctx, batchSize)
		if err != nil || len(works) == 0 {
			return embedded, err
		}
		texts := make([]string, len(works))
		for i := range works {
			texts[i] = works[i].Passage()
		}
		vectors, err := e.embedder.EmbedPassages(ctx, texts)
		if err != nil {
			return embedded, err
		}
		for i, vector := range vectors {
			if err := e.catalog.SaveEmbedding(ctx, works[i].Key, vector); err != nil {
				return embedded, err
			}
		}
		embedded += len(works)
	}
	return embedded, nil
}
