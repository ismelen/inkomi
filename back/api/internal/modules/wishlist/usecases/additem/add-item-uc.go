package additem

import (
	"context"
	"log"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

const (
	maxItems = 500
	// resolveTimeout bounds the catalog lookup of a new book; a slow one is left to the backfill.
	resolveTimeout = 6 * time.Second
)

type AddItemUC struct {
	repo     ports.IItemRepository
	resolver *resolvework.ResolveWorkUC
}

func NewAddItemUC(repo ports.IItemRepository, resolver *resolvework.ResolveWorkUC) *AddItemUC {
	return &AddItemUC{repo: repo, resolver: resolver}
}

// Execute stores the item and, for a book, looks up its cover and synopsis; coverURL is the
// cover of the suggestion the user picked, if any. A failed lookup doesn't fail the add.
func (a *AddItemUC) Execute(ctx context.Context, userId int, kind models.ItemKind, title, author, workKey, authorKey, coverURL string) (*models.Item, error) {
	item, err := models.NewItem(userId, kind, title, author, workKey, authorKey)
	if err != nil {
		return nil, err
	}
	count, err := a.repo.CountByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if count >= maxItems {
		return nil, models.ErrWishlistFull
	}
	if err := a.repo.Add(ctx, item); err != nil {
		return nil, err
	}

	resolveCtx, cancel := context.WithTimeout(ctx, resolveTimeout)
	defer cancel()
	if err := a.resolver.Execute(resolveCtx, item, coverURL); err != nil {
		log.Printf("wishlist: resolve work of item %d: %v", item.Id, err)
	}
	return item, nil
}
