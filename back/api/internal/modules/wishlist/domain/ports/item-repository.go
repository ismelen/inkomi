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
	// SetWorkKey stores the work key found for a book typed by hand; it returns
	// models.ErrDuplicate when the user already has that work.
	SetWorkKey(ctx context.Context, item *models.Item, workKey string) error
	// ListBooksWithoutWork returns the books of every user with no work key or whose work has
	// no stored details.
	ListBooksWithoutWork(ctx context.Context) ([]models.Item, error)
}
