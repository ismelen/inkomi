package wishlistreader

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listallitems"
)

// WishlistReader reads the wishlists from the wishlist module, in process.
type WishlistReader struct {
	listUC *listallitems.ListAllItemsUC
}

var _ ports.IWishlist = (*WishlistReader)(nil)

func NewWishlistReader(listUC *listallitems.ListAllItemsUC) *WishlistReader {
	return &WishlistReader{listUC: listUC}
}

func (w *WishlistReader) ListAll(ctx context.Context) ([]models.WishedItem, error) {
	items, err := w.listUC.Execute(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]models.WishedItem, len(items))
	for i, item := range items {
		result[i] = models.WishedItem{
			Id:        item.Id,
			UserId:    item.UserId,
			Kind:      models.SignalKind(item.Kind),
			Title:     item.Title,
			Author:    item.Author,
			WorkKey:   item.WorkKey,
			AuthorKey: item.AuthorKey,
		}
	}
	return result, nil
}
