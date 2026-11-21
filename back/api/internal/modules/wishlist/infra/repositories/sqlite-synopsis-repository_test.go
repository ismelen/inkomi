package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
)

func TestSQLiteSynopsisRepository_SaveGet_KeepsOnePerLanguageAndReplaces(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteSynopsisRepository(setupDB(t))

	// Act
	_, missingErr := repo.Get(ctx, "OL5W", "es")
	require.NoError(t, repo.Save(ctx, models.Synopsis{WorkKey: "OL5W", Language: "es", Text: "Cae ceniza."}))
	require.NoError(t, repo.Save(ctx, models.Synopsis{WorkKey: "OL5W", Language: "fr", Text: ""}))
	require.NoError(t, repo.Save(ctx, models.Synopsis{WorkKey: "OL5W", Language: "es", Text: "Actualizada."}))
	es, esErr := repo.Get(ctx, "OL5W", "es")
	fr, frErr := repo.Get(ctx, "OL5W", "fr")

	// Assert
	assert.ErrorIs(t, missingErr, models.ErrSynopsisNotFound)
	require.NoError(t, esErr)
	assert.Equal(t, "Actualizada.", es.Text)
	assert.WithinDuration(t, time.Now(), es.FetchedAt, time.Minute)
	require.NoError(t, frErr)
	assert.Empty(t, fr.Text)
}
