package getdropboxaccess

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

// accessMargin is how long a cached access token must still be valid to be handed out, so a
// long upload doesn't run out of it.
const accessMargin = 5 * time.Minute

type GetDropboxAccessUC struct {
	links ports.IDropboxLinkRepository
	oauth ports.IDropboxOAuth
}

func NewGetDropboxAccessUC(links ports.IDropboxLinkRepository, oauth ports.IDropboxOAuth) *GetDropboxAccessUC {
	return &GetDropboxAccessUC{links: links, oauth: oauth}
}

// Execute returns the Dropbox link of the user with a valid access token, refreshing it when
// needed. A revoked refresh token unlinks Dropbox and returns models.ErrDropboxNotLinked.
func (g *GetDropboxAccessUC) Execute(ctx context.Context, userId int) (*models.DropboxLink, error) {
	link, err := g.links.Get(ctx, userId)
	if err != nil {
		return nil, err
	}
	if link.HasValidAccess(time.Now(), accessMargin) {
		return link, nil
	}

	access, err := g.oauth.Refresh(ctx, link.RefreshToken)
	if errors.Is(err, models.ErrInvalidExternalToken) {
		err = fmt.Errorf("%w: %w", models.ErrDropboxNotLinked, err)
		return nil, errors.Join(err, g.links.Delete(ctx, userId))
	}
	if err != nil {
		return nil, err
	}

	if err := g.links.UpdateAccess(ctx, userId, access); err != nil {
		return nil, err
	}
	link.Access = access
	return link, nil
}
