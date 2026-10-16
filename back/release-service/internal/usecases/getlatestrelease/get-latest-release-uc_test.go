package getlatestrelease_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/getlatestrelease"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sourceReturning(release *models.Release, err error) *mocks.MockReleaseSource {
	return &mocks.MockReleaseSource{
		LatestFn: func(ctx context.Context) (*models.Release, error) { return release, err },
	}
}

func TestGetLatestReleaseUC_Execute_WithAPK_ReturnsRelease(t *testing.T) {
	t.Parallel()

	// Arrange
	release := &models.Release{
		Version: models.Version{Major: 2},
		APK:     models.ReleaseAsset{Name: "inkomi.apk", URL: "https://github.com/x/inkomi.apk"},
	}
	uc := getlatestrelease.NewGetLatestReleaseUC(sourceReturning(release, nil))

	// Act
	got, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, release, got)
}

func TestGetLatestReleaseUC_Execute_NoAPKURL_ReturnsAPKNotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getlatestrelease.NewGetLatestReleaseUC(sourceReturning(&models.Release{}, nil))

	// Act
	got, err := uc.Execute(context.Background())

	// Assert
	assert.ErrorIs(t, err, models.ErrAPKNotFound)
	assert.Nil(t, got)
}

func TestGetLatestReleaseUC_Execute_SourceFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getlatestrelease.NewGetLatestReleaseUC(&mocks.MockReleaseSource{})

	// Act
	got, err := uc.Execute(context.Background())

	// Assert
	assert.ErrorIs(t, err, models.ErrReleaseNotFound)
	assert.Nil(t, got)
}
