package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type IWishlist interface {
	// ListAll returns the wishlist items of every user.
	ListAll(ctx context.Context) ([]models.WishedItem, error)
}
