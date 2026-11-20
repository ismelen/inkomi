package removeitem

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

type RemoveItemUC struct {
	repo     ports.IItemRepository
	onChange func(userId int)
}

func NewRemoveItemUC(repo ports.IItemRepository) *RemoveItemUC {
	return &RemoveItemUC{repo: repo}
}

// SetOnChange registers f to be called with the user after an item is removed.
func (r *RemoveItemUC) SetOnChange(f func(userId int)) {
	r.onChange = f
}

func (r *RemoveItemUC) Execute(ctx context.Context, userId int, id int64) error {
	if err := r.repo.Delete(ctx, id, userId); err != nil {
		return err
	}
	if r.onChange != nil {
		r.onChange(userId)
	}
	return nil
}
