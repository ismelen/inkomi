package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/repositories"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tree creates a joined folder with two chapters in status and returns the ids.
func tree(t *testing.T, db *sql.DB, status models.SourceStatus) (root string, children []string) {
	t.Helper()
	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)

	root, err := repo.Create(ctx, &models.Source{UserId: 1, Title: "Folder", Filename: "Folder", Join: true, IsFile: false})
	require.NoError(t, err)
	for i := range 2 {
		size := int64(10)
		id, err := repo.Create(ctx, &models.Source{UserId: 1, Title: "Ch", Filename: "ch.cbz", Size: &size, IsFile: true, ParentId: &root, Position: i})
		require.NoError(t, err)
		if status != models.SourceStatusPendingUpload {
			require.NoError(t, repo.UpdateStatus(ctx, id, 1, status))
		}
		children = append(children, id)
	}
	return root, children
}

func TestSQLiteStorageReservationRepository_CreateListRelease(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()
	repo := repositories.NewSQLiteStorageReservationRepository(db)
	ctx := context.Background()

	// Act
	require.NoError(t, repo.Create(ctx, &models.StorageReservation{RootId: "a", UserId: 1, Bytes: 100}))
	require.NoError(t, repo.Create(ctx, &models.StorageReservation{RootId: "b", UserId: 1, Bytes: 50}))
	require.NoError(t, repo.Release(ctx, "a"))
	active, err := repo.ListActive(ctx)

	// Assert
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "b", active[0].RootId)
	assert.Equal(t, int64(50), active[0].Bytes)
}

func TestSQLiteStorageReservationRepository_TreeQueries(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()
	repo := repositories.NewSQLiteStorageReservationRepository(db)
	ctx := context.Background()
	pendingRoot, pendingChildren := tree(t, db, models.SourceStatusPendingUpload)
	joinedRoot, joinedChildren := tree(t, db, models.SourceStatusJoined)

	// Act & Assert
	root, err := repo.RootOf(ctx, pendingChildren[0], 1)
	require.NoError(t, err)
	assert.Equal(t, pendingRoot, root)
	root, err = repo.RootOf(ctx, pendingRoot, 1)
	require.NoError(t, err)
	assert.Equal(t, pendingRoot, root)

	ids, err := repo.TreeIds(ctx, pendingRoot)
	require.NoError(t, err)
	assert.ElementsMatch(t, append([]string{pendingRoot}, pendingChildren...), ids)

	unconfirmed, err := repo.IsTreeUnconfirmed(ctx, pendingRoot)
	require.NoError(t, err)
	assert.True(t, unconfirmed)
	unconfirmed, err = repo.IsTreeUnconfirmed(ctx, joinedRoot)
	require.NoError(t, err)
	assert.False(t, unconfirmed)

	finished, err := repo.IsTreeFinished(ctx, joinedRoot)
	require.NoError(t, err)
	assert.False(t, finished, "a joined folder needs its results")

	sources := repositories.NewSQLiteSourceRepository(db)
	result := models.NewResultSource(resultMsg(joinedRoot), models.SourceStatusSent)
	require.NoError(t, sources.UpsertResult(ctx, result))
	finished, err = repo.IsTreeFinished(ctx, joinedRoot)
	require.NoError(t, err)
	assert.True(t, finished)
	_ = joinedChildren
}

func resultMsg(parentId string) jobs.Msg {
	return jobs.Msg{
		Id:       "result-1",
		UserId:   1,
		Title:    "Folder Ch.[1 - 2]",
		Filename: "Folder.epub",
		Result:   &jobs.Result{ParentId: parentId, Index: 0, Total: 1},
	}
}
