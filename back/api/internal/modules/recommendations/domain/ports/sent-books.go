package ports

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type ISentBooks interface {
	// ListSince returns up to limit books sent at or after since, oldest first.
	ListSince(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error)
}
