package embedworks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/embedworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
)

func TestEmbedWorksUC_Execute_Unembedded_SavesVectorsAndReleasesTheModel(t *testing.T) {
	t.Parallel()

	// Arrange
	queue := [][]models.Work{
		{{Key: "OL1W", Title: "La nena"}, {Key: "OL2W", Title: "Dune"}},
		nil,
	}
	catalog := &mocks.MockCatalogRepository{}
	catalog.ListUnembeddedFn = func(ctx context.Context, limit int) ([]models.Work, error) {
		batch := queue[0]
		queue = queue[1:]
		return batch, nil
	}
	saved := map[string]models.Embedding{}
	catalog.SaveEmbeddingFn = func(ctx context.Context, key string, e models.Embedding) error {
		saved[key] = e
		return nil
	}
	var texts []string
	embedder := &mocks.MockEmbedder{EmbedPassagesFn: func(ctx context.Context, in []string) ([]models.Embedding, error) {
		texts = in
		return []models.Embedding{{1, 0}, {0, 1}}, nil
	}}
	uc := embedworks.NewEmbedWorksUC(embedder, catalog, func() bool { return true })

	// Act
	embedded, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, embedded)
	assert.Equal(t, []string{"La nena", "Dune"}, texts)
	assert.Equal(t, models.Embedding{0, 1}, saved["OL2W"])
	assert.Equal(t, 1, embedder.Releases)
}

func TestEmbedWorksUC_Execute_LowMemoryOrNoModel_StopsAndReleases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		ready bool
		err   error
	}{
		{"Low memory", false, nil},
		{"No model", true, models.ErrNoModel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			catalog := &mocks.MockCatalogRepository{ListUnembeddedFn: func(ctx context.Context, limit int) ([]models.Work, error) {
				return []models.Work{{Key: "OL1W"}}, nil
			}}
			embedder := &mocks.MockEmbedder{EmbedPassagesFn: func(ctx context.Context, in []string) ([]models.Embedding, error) {
				return nil, tc.err
			}}
			uc := embedworks.NewEmbedWorksUC(embedder, catalog, func() bool { return tc.ready })

			// Act
			embedded, err := uc.Execute(context.Background())

			// Assert
			assert.Zero(t, embedded)
			assert.True(t, errors.Is(err, tc.err) || (tc.err == nil && err == nil))
			assert.Equal(t, 1, embedder.Releases)
		})
	}
}
