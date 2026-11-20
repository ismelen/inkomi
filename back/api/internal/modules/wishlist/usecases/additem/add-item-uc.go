package additem

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const maxItems = 500

type AddItemUC struct {
	repo     ports.IItemRepository
	onChange func(userId int)
}

func NewAddItemUC(repo ports.IItemRepository) *AddItemUC {
	return &AddItemUC{repo: repo}
}

// SetOnChange registers f to be called with the user after an item is added.
func (a *AddItemUC) SetOnChange(f func(userId int)) {
	a.onChange = f
}

func (a *AddItemUC) Execute(ctx context.Context, userId int, kind models.ItemKind, title, author, workKey, authorKey string) (*models.Item, error) {
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
	if a.onChange != nil {
		a.onChange(userId)
	}
	return item, nil
}
