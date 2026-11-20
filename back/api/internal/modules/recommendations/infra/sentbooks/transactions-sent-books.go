package sentbooks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/usecases/listsentbooks"
)

// TransactionsSentBooks reads the sent books from the transactions module, in process.
type TransactionsSentBooks struct {
	listUC *listsentbooks.ListSentBooksUC
}

var _ ports.ISentBooks = (*TransactionsSentBooks)(nil)

func NewTransactionsSentBooks(listUC *listsentbooks.ListSentBooksUC) *TransactionsSentBooks {
	return &TransactionsSentBooks{listUC: listUC}
}

func (t *TransactionsSentBooks) ListSince(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error) {
	books, err := t.listUC.Execute(ctx, since, limit)
	if err != nil {
		return nil, err
	}
	result := make([]models.SentBook, len(books))
	for i, b := range books {
		result[i] = models.SentBook{SourceId: b.SourceId, UserId: b.UserId, Title: b.Title, Filename: b.Filename, UpdatedAt: b.UpdatedAt}
	}
	return result, nil
}
