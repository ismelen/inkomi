package listallitems

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

// ListAllItemsUC lists the wishlists of every user, for the recommendations module.
type ListAllItemsUC struct {
	repo ports.IItemRepository
}

func NewListAllItemsUC(repo ports.IItemRepository) *ListAllItemsUC {
	return &ListAllItemsUC{repo: repo}
}

func (l *ListAllItemsUC) Execute(ctx context.Context) ([]models.Item, error) {
	return l.repo.ListAll(ctx)
}
