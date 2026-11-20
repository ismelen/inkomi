package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type IItemRepository interface {
	// Add stores item and sets its Id and CreatedAt; it returns models.ErrDuplicate when the user
	// already has the same item.
	Add(ctx context.Context, item *models.Item) error
	// Delete returns models.ErrItemNotFound when the user has no item with id.
	Delete(ctx context.Context, id int64, userId int) error
	// ListByUserId returns the items of the user, newest first.
	ListByUserId(ctx context.Context, userId int) ([]models.Item, error)
	CountByUserId(ctx context.Context, userId int) (int, error)
	// ListAll returns every item of every user, oldest first.
	ListAll(ctx context.Context) ([]models.Item, error)
}
