package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type IDropboxLinkRepository interface {
	Get(ctx context.Context, userId int) (*models.DropboxLink, error)
	// Save stores the account and tokens of link, keeping the folder of an existing link.
	Save(ctx context.Context, link *models.DropboxLink) error
	UpdateAccess(ctx context.Context, userId int, access models.AccessToken) error
	UpdateFolder(ctx context.Context, userId int, folder string) error
	Delete(ctx context.Context, userId int) error
}
