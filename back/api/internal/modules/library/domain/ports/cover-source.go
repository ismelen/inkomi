package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
)

type ICoverSource interface {
	// Cover returns models.ErrInvalidCoverURL when url isn't a cover of the library.
	Cover(ctx context.Context, url string) (*models.Cover, error)
}
