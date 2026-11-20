package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type IEmbedder interface {
	// EmbedPassages returns the normalized embedding of each text, in order.
	EmbedPassages(ctx context.Context, texts []string) ([]models.Embedding, error)
	// Release frees the model until the next call: it is big and only needed during a refresh.
	Release()
}
