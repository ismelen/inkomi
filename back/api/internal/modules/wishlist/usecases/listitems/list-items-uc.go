package listitems

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

type ListItemsUC struct {
	repo ports.IItemRepository
}

func NewListItemsUC(repo ports.IItemRepository) *ListItemsUC {
	return &ListItemsUC{repo: repo}
}

// Execute returns the items of the user, newest first.
func (l *ListItemsUC) Execute(ctx context.Context, userId int) ([]models.Item, error) {
	return l.repo.ListByUserId(ctx, userId)
}
