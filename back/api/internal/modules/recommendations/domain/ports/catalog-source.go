package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

// ICatalogSource is where works and their metadata come from (Open Library for books).
type ICatalogSource interface {
	// MatchBook returns the work that best matches title (and author when known), or nil when
	// nothing is close enough.
	MatchBook(ctx context.Context, title, author string) (*models.Work, error)
	// Work returns the work with key, or nil when it doesn't exist.
	Work(ctx context.Context, key string) (*models.Work, error)
	// MatchAuthor returns the key and canonical name of the author, or "" when not found.
	MatchAuthor(ctx context.Context, name string) (key, canonical string, err error)
	Description(ctx context.Context, workKey string) (string, error)
	WorksByAuthor(ctx context.Context, authorKey string, limit int) ([]models.Work, error)
	WorksBySubject(ctx context.Context, subject string, limit int) ([]models.Work, error)
	// NewWorks lists works first published in sinceYear or later, of a subject or of an author
	// (one of them is set).
	NewWorks(ctx context.Context, subject, authorKey string, sinceYear, limit int) ([]models.Work, error)
}
