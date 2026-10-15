package dropboxlogin

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
)

type DropboxLoginUC struct {
	oauth        ports.IDropboxOAuth
	links        ports.IDropboxLinkRepository
	signIn       *socialsignin.SocialSignInUC
	startSession *startsession.StartSessionUC
}

func NewDropboxLoginUC(
	oauth ports.IDropboxOAuth,
	links ports.IDropboxLinkRepository,
	signIn *socialsignin.SocialSignInUC,
	startSession *startsession.StartSessionUC,
) *DropboxLoginUC {
	return &DropboxLoginUC{oauth: oauth, links: links, signIn: signIn, startSession: startSession}
}

// Execute signs in with the Dropbox account of refreshToken, which also becomes the account
// where the files of the user are uploaded.
func (d *DropboxLoginUC) Execute(ctx context.Context, refreshToken, deviceId string) (*models.TokenPair, error) {
	if err := models.ValidateDeviceId(deviceId); err != nil {
		return nil, err
	}

	access, err := d.oauth.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	account, err := d.oauth.Account(ctx, access.Token)
	if err != nil {
		return nil, err
	}

	user, err := d.signIn.Execute(ctx, *account)
	if err != nil {
		return nil, err
	}

	if err := d.links.Save(ctx, models.NewDropboxLink(user.Id, *account, refreshToken, access)); err != nil {
		return nil, err
	}
	return d.startSession.Execute(ctx, user, deviceId)
}
