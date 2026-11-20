package getcover_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/getcover"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/mocks"
)

func TestGetCoverUC_Execute_EmptyURL_ReturnsErrInvalidCoverURL(t *testing.T) {
	t.Parallel()

	// Arrange
	called := false
	uc := getcover.NewGetCoverUC(&mocks.MockCoverSource{
		CoverFn: func(ctx context.Context, url string) (*models.Cover, error) {
			called = true
			return nil, nil
		},
	})

	// Act
	cover, err := uc.Execute(context.Background(), "  ")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidCoverURL)
	assert.Nil(t, cover)
	assert.False(t, called)
}

func TestGetCoverUC_Execute_URL_ReturnsSourceCover(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotURL string
	uc := getcover.NewGetCoverUC(&mocks.MockCoverSource{
		CoverFn: func(ctx context.Context, url string) (*models.Cover, error) {
			gotURL = url
			return &models.Cover{ContentType: "image/jpeg"}, nil
		},
	})

	// Act
	cover, err := uc.Execute(context.Background(), " https://libgen.la/covers/1/a_small.jpg ")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "https://libgen.la/covers/1/a_small.jpg", gotURL)
	assert.Equal(t, "image/jpeg", cover.ContentType)
}
