package listsentbooks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/ports"
)

const maxLimit = 500

// ListSentBooksUC lists the books users sent (library books and epubs, not manga) so other
// modules can follow them in process.
type ListSentBooksUC struct {
	sourceRepo ports.ISourceRepository
}

func NewListSentBooksUC(sourceRepo ports.ISourceRepository) *ListSentBooksUC {
	return &ListSentBooksUC{sourceRepo: sourceRepo}
}

// Execute returns up to limit books updated at or after since, oldest first.
func (l *ListSentBooksUC) Execute(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error) {
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}
	return l.sourceRepo.ListSentBooks(ctx, since, limit)
}
