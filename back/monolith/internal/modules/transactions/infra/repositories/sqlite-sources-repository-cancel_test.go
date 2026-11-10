package repositories_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteSourceRepository_CancelTree_CancelsOnlyUnfinished(t *testing.T) {
	// Arrange: a folder whose first chapter was already sent
	db := setupTestDB(t)
	defer db.Close()
	repo := repositories.NewSQLiteSourceRepository(db)
	ctx := context.Background()
	root, children := tree(t, db, models.SourceStatusProcessing)
	require.NoError(t, repo.UpdateStatus(ctx, children[0], 1, models.SourceStatusSent))

	// Act
	ids, err := repo.CancelTree(ctx, root, 1)

	// Assert
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{root, children[1]}, ids)
	sent, err := repo.GetByIdAndUserId(ctx, children[0], 1)
	require.NoError(t, err)
	assert.Equal(t, models.SourceStatusSent, sent.Status)
	cancelled, err := repo.GetByIdAndUserId(ctx, children[1], 1)
	require.NoError(t, err)
	assert.Equal(t, models.SourceStatusCancelled, cancelled.Status)
	assert.NotNil(t, cancelled.CompletedAt)
}

func TestSQLiteSourceRepository_LateEvents_DontOverwriteCancelled(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()
	repo := repositories.NewSQLiteSourceRepository(db)
	ctx := context.Background()
	root, children := tree(t, db, models.SourceStatusProcessing)
	_, err := repo.CancelTree(ctx, root, 1)
	require.NoError(t, err)

	// Act: a job that was already running reports back
	require.NoError(t, repo.UpdateStatus(ctx, children[0], 1, models.SourceStatusJoining))
	require.NoError(t, repo.UpdateErrorMsg(ctx, children[1], 1, "interrupted"))

	// Assert
	for _, id := range children {
		src, err := repo.GetByIdAndUserId(ctx, id, 1)
		require.NoError(t, err)
		assert.Equal(t, models.SourceStatusCancelled, src.Status, id)
	}
}
