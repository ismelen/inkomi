package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/domain/models"
)

type IUserSettings interface {
	// GetDropboxAccount returns models.ErrDropboxNotLinked when the user has no Dropbox account.
	GetDropboxAccount(ctx context.Context, userId int) (*models.DropboxAccount, error)
	// GetFCMToken returns models.ErrFCMTokenNotFound when the user has no device registered.
	GetFCMToken(ctx context.Context, userId int) (string, error)
}
