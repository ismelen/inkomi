package getsample_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/getsample"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSampleUC_Execute_Sample_ReturnsSourceBytes(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getsample.NewGetSampleUC(&mocks.MockSampleSource{})

	// Act
	data, err := uc.Execute(context.Background(), models.PreviewSampleColor)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "color", string(data))
}
