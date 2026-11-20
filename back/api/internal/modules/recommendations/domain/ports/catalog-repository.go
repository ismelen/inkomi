package ports

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type ICatalogRepository interface {
	// UpsertWork stores or refreshes a work and marks it as seen now. A user work never goes back
	// to candidate, and the embedding is dropped when the passage changed.
	UpsertWork(ctx context.Context, work *models.Work) (created bool, err error)
	GetWorks(ctx context.Context, keys []string) (map[string]models.Work, error)
	// ListUnembedded returns up to limit works without embedding, user works first.
	ListUnembedded(ctx context.Context, limit int) ([]models.Work, error)
	SaveEmbedding(ctx context.Context, key string, embedding models.Embedding) error
	// ListVectors returns every embedded work.
	ListVectors(ctx context.Context) ([]models.WorkVector, error)
	// PruneCandidates deletes the candidates not seen since before that no user has a signal
	// for, and returns how many went.
	PruneCandidates(ctx context.Context, before time.Time) (int, error)
	// TopSubjects returns the subjects most common among the users' works.
	TopSubjects(ctx context.Context, limit int) ([]string, error)
	// TopAuthorKeys returns the authors most common among the users' works and wished authors.
	TopAuthorKeys(ctx context.Context, limit int) ([]string, error)
}
