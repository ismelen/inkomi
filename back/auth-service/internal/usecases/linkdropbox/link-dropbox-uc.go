package linkdropbox

import (
	"context"
	"errors"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type LinkDropboxUC struct {
	oauth      ports.IDropboxOAuth
	identities ports.IIdentityRepository
	links      ports.IDropboxLinkRepository
}

func NewLinkDropboxUC(oauth ports.IDropboxOAuth, identities ports.IIdentityRepository, links ports.IDropboxLinkRepository) *LinkDropboxUC {
	return &LinkDropboxUC{oauth: oauth, identities: identities, links: links}
}

// Execute sets the Dropbox account of refreshToken as the one where the files of the user are
// uploaded, and lets the user sign in with it unless it already belongs to another user.
func (l *LinkDropboxUC) Execute(ctx context.Context, userId int, refreshToken string) error {
	access, err := l.oauth.Refresh(ctx, refreshToken)
	if err != nil {
		return err
	}
	account, err := l.oauth.Account(ctx, access.Token)
	if err != nil {
		return err
	}

	identity, err := l.identities.Get(ctx, models.ProviderDropbox, account.Subject)
	switch {
	case errors.Is(err, models.ErrIdentityNotFound):
		identity := models.Identity{Provider: models.ProviderDropbox, Subject: account.Subject, UserId: userId}
		if err := l.identities.Create(ctx, identity); err != nil {
			return err
		}
	case err != nil:
		return err
	case identity.UserId != userId:
		return models.ErrDropboxAccountTaken
	}

	return l.links.Save(ctx, models.NewDropboxLink(userId, *account, refreshToken, access))
}
