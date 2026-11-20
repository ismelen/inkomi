package removeitem

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

type RemoveItemUC struct {
	repo ports.IItemRepository
}

func NewRemoveItemUC(repo ports.IItemRepository) *RemoveItemUC {
	return &RemoveItemUC{repo: repo}
}

func (r *RemoveItemUC) Execute(ctx context.Context, userId int, id int64) error {
	return r.repo.Delete(ctx, id, userId)
}
