package getprofile

import (
	"context"
	"errors"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type GetProfileUC struct {
	users      ports.IUserRepository
	identities ports.IIdentityRepository
	links      ports.IDropboxLinkRepository
}

func NewGetProfileUC(users ports.IUserRepository, identities ports.IIdentityRepository, links ports.IDropboxLinkRepository) *GetProfileUC {
	return &GetProfileUC{users: users, identities: identities, links: links}
}

func (g *GetProfileUC) Execute(ctx context.Context, userId int) (*models.Profile, error) {
	user, err := g.users.GetById(ctx, userId)
	if err != nil {
		return nil, err
	}

	providers, err := g.identities.ListProviders(ctx, userId)
	if err != nil {
		return nil, err
	}

	link, err := g.links.Get(ctx, userId)
	if err != nil && !errors.Is(err, models.ErrDropboxNotLinked) {
		return nil, err
	}

	return &models.Profile{User: user, Providers: providers, Dropbox: link}, nil
}
