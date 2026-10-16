package getlatestrelease

import (
	"context"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/domain/ports"
)

type GetLatestReleaseUC struct {
	source ports.IReleaseSource
}

func NewGetLatestReleaseUC(source ports.IReleaseSource) *GetLatestReleaseUC {
	return &GetLatestReleaseUC{source: source}
}

func (u *GetLatestReleaseUC) Execute(ctx context.Context) (*models.Release, error) {
	release, err := u.source.Latest(ctx)
	if err != nil {
		return nil, err
	}
	if release.APK.URL == "" {
		return nil, models.ErrAPKNotFound
	}
	return release, nil
}
