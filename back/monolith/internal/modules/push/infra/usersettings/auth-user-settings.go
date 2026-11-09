package usersettings

import (
	"context"
	"errors"

	authmodels "github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/domain/ports"
)

var _ ports.IUserSettings = (*AuthUserSettings)(nil)

type IGetDropboxAccess interface {
	Execute(ctx context.Context, userId int) (*authmodels.DropboxLink, error)
}

type IGetFCMToken interface {
	Execute(ctx context.Context, userId int) (string, error)
}

// AuthUserSettings reads the user settings straight from the use cases of the auth module.
type AuthUserSettings struct {
	getDropboxAccess IGetDropboxAccess
	getFCMToken      IGetFCMToken
}

func NewAuthUserSettings(getDropboxAccess IGetDropboxAccess, getFCMToken IGetFCMToken) *AuthUserSettings {
	return &AuthUserSettings{getDropboxAccess: getDropboxAccess, getFCMToken: getFCMToken}
}

func (a *AuthUserSettings) GetDropboxAccount(ctx context.Context, userId int) (*models.DropboxAccount, error) {
	link, err := a.getDropboxAccess.Execute(ctx, userId)
	if errors.Is(err, authmodels.ErrDropboxNotLinked) || errors.Is(err, authmodels.ErrUserNotFound) {
		return nil, models.ErrDropboxNotLinked
	}
	if err != nil {
		return nil, err
	}
	return models.NewDropboxAccount(link.Access.Token, link.Folder)
}

func (a *AuthUserSettings) GetFCMToken(ctx context.Context, userId int) (string, error) {
	token, err := a.getFCMToken.Execute(ctx, userId)
	if errors.Is(err, authmodels.ErrFCMTokenNotFound) || errors.Is(err, authmodels.ErrUserNotFound) {
		return "", models.ErrFCMTokenNotFound
	}
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", models.ErrFCMTokenNotFound
	}
	return token, nil
}
