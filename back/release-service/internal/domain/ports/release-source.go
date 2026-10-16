package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
)

type IReleaseSource interface {
	// Latest returns models.ErrReleaseNotFound when nothing is published and
	// models.ErrAPKNotFound when the latest release doesn't carry the apk.
	Latest(ctx context.Context) (*models.Release, error)
}
