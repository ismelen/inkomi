package repositories_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var listBase = time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC)

// createAt creates a source and pins its createdAt and updatedAt to at.
func createAt(t *testing.T, db *sql.DB, repo ports.ISourceRepository, source *models.Source, at time.Time) string {
	t.Helper()

	id, err := repo.Create(context.Background(), source)
	require.NoError(t, err)
	setTimes(t, db, id, at)
	return id
}

func setTimes(t *testing.T, db *sql.DB, id string, at time.Time) {
	t.Helper()

	stamp := at.UTC().Format("2006-01-02 15:04:05")
	_, err := db.Exec(`UPDATE source SET createdAt = ?, updatedAt = ? WHERE id = ?`, stamp, stamp, id)
	require.NoError(t, err)
}

func setStatus(t *testing.T, db *sql.DB, id string, status models.SourceStatus) {
	t.Helper()

	_, err := db.Exec(`UPDATE source SET status = ? WHERE id = ?`, status, id)
	require.NoError(t, err)
}

func queuedFile(t *testing.T, db *sql.DB, repo ports.ISourceRepository, userId int, title string, at time.Time) string {
	t.Helper()

	id := createAt(t, db, repo, &models.Source{UserId: userId, Title: title, Filename: title + ".cbz", IsFile: true}, at)
	setStatus(t, db, id, models.SourceStatusQueued)
	return id
}

func ids(trees []models.SourceTree) []string {
	result := make([]string, len(trees))
	for i := range trees {
		result[i] = trees[i].Id
	}
	return result
}

func TestSQLiteSourceRepository_ListTreesByUserId_Paginated_ReturnsNewestFirst(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	oldest := queuedFile(t, db, repo, 1, "a", listBase)
	middle := queuedFile(t, db, repo, 1, "b", listBase.Add(time.Minute))
	newest := queuedFile(t, db, repo, 1, "c", listBase.Add(2*time.Minute))

	// Act
	first, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 2})
	require.NoError(t, err)
	cursor := models.NewSourceCursor(&first[1].Source)
	second, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 2, Cursor: cursor})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{newest, middle}, ids(first))
	assert.Equal(t, []string{oldest}, ids(second))
}

func TestSQLiteSourceRepository_ListTreesByUserId_SameCreatedAt_PaginatesById(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	for _, title := range []string{"a", "b", "c"} {
		queuedFile(t, db, repo, 1, title, listBase)
	}

	// Act
	first, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 2})
	require.NoError(t, err)
	second, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 2, Cursor: models.NewSourceCursor(&first[1].Source)})

	// Assert
	require.NoError(t, err)
	require.Len(t, first, 2)
	require.Len(t, second, 1)
	assert.Greater(t, first[0].Id, first[1].Id)
	assert.Greater(t, first[1].Id, second[0].Id)
}

func TestSQLiteSourceRepository_ListTreesByUserId_Folder_ReturnsChildrenAndResults(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	folderId := createAt(t, db, repo, &models.Source{UserId: 1, Title: "Folder", Join: true}, listBase)
	second := createAt(t, db, repo, &models.Source{UserId: 1, Title: "Ch2", Filename: "2.cbz", IsFile: true, ParentId: &folderId, Position: 1}, listBase)
	first := createAt(t, db, repo, &models.Source{UserId: 1, Title: "Ch1", Filename: "1.cbz", IsFile: true, ParentId: &folderId, Position: 0}, listBase)
	setStatus(t, db, first, models.SourceStatusJoined)
	setStatus(t, db, second, models.SourceStatusJoined)
	require.NoError(t, repo.UpsertResult(ctx, resultSource(folderId, 0, models.SourceStatusSent)))

	// Act
	trees, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 10})

	// Assert
	require.NoError(t, err)
	require.Len(t, trees, 1)
	assert.Equal(t, folderId, trees[0].Id)
	require.Len(t, trees[0].Children, 3)
	assert.Equal(t, first, trees[0].Children[0].Id)
	assert.Equal(t, second, trees[0].Children[1].Id)
	assert.True(t, trees[0].Children[2].IsResult)
}

func TestSQLiteSourceRepository_ListTreesByUserId_NeverConfirmed_SkipsTree(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	createAt(t, db, repo, &models.Source{UserId: 1, Title: "abandoned", Filename: "a.cbz", IsFile: true}, listBase)
	abandonedFolder := createAt(t, db, repo, &models.Source{UserId: 1, Title: "abandoned folder"}, listBase)
	createAt(t, db, repo, &models.Source{UserId: 1, Title: "x", Filename: "x.cbz", IsFile: true, ParentId: &abandonedFolder}, listBase)
	folder := createAt(t, db, repo, &models.Source{UserId: 1, Title: "folder"}, listBase)
	child := createAt(t, db, repo, &models.Source{UserId: 1, Title: "y", Filename: "y.cbz", IsFile: true, ParentId: &folder}, listBase)
	setStatus(t, db, child, models.SourceStatusQueued)

	// Act
	trees, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 10})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{folder}, ids(trees))
}

func TestSQLiteSourceRepository_ListTreesByUserId_UpdatedSince_ReturnsTreesWithAnyUpdatedNode(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	since := listBase.Add(time.Hour)
	queuedFile(t, db, repo, 1, "stale", listBase)
	updatedRoot := queuedFile(t, db, repo, 1, "root", listBase)
	setTimes(t, db, updatedRoot, since)
	folder := createAt(t, db, repo, &models.Source{UserId: 1, Title: "folder"}, listBase)
	child := createAt(t, db, repo, &models.Source{UserId: 1, Title: "c", Filename: "c.cbz", IsFile: true, ParentId: &folder}, listBase)
	setStatus(t, db, child, models.SourceStatusProcessing)
	_, err := db.Exec(`UPDATE source SET updatedAt = ? WHERE id = ?`, since.Add(time.Second).Format("2006-01-02 15:04:05"), child)
	require.NoError(t, err)

	// Act
	trees, err := repo.ListTreesByUserId(ctx, 1, models.SourceListFilter{Limit: 10, UpdatedSince: &since})

	// Assert
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{updatedRoot, folder}, ids(trees))
}

func TestSQLiteSourceRepository_ListTreesByUserId_OtherUser_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	queuedFile(t, db, repo, 1, "mine", listBase)

	// Act
	trees, err := repo.ListTreesByUserId(context.Background(), 2, models.SourceListFilter{Limit: 10})

	// Assert
	require.NoError(t, err)
	assert.Empty(t, trees)
}

func TestSQLiteSourceRepository_UpdateStatus_Terminal_SetsCompletedAt(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	processing, err := repo.Create(ctx, &models.Source{UserId: 1, Filename: "a.cbz", IsFile: true})
	require.NoError(t, err)
	sent, err := repo.Create(ctx, &models.Source{UserId: 1, Filename: "b.cbz", IsFile: true})
	require.NoError(t, err)
	failed, err := repo.Create(ctx, &models.Source{UserId: 1, Filename: "c.cbz", IsFile: true})
	require.NoError(t, err)

	// Act
	require.NoError(t, repo.UpdateStatus(ctx, processing, 1, models.SourceStatusProcessing))
	require.NoError(t, repo.UpdateStatus(ctx, sent, 1, models.SourceStatusSent))
	require.NoError(t, repo.UpdateErrorMsg(ctx, failed, 1, "boom"))

	// Assert
	got, err := repo.GetByIdAndUserId(ctx, processing, 1)
	require.NoError(t, err)
	assert.Nil(t, got.CompletedAt)
	got, err = repo.GetByIdAndUserId(ctx, sent, 1)
	require.NoError(t, err)
	assert.NotNil(t, got.CompletedAt)
	got, err = repo.GetByIdAndUserId(ctx, failed, 1)
	require.NoError(t, err)
	assert.NotNil(t, got.CompletedAt)
}

func TestSQLiteSourceRepository_Now_ReturnsDatabaseClock(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)

	// Act
	now, err := repo.Now(context.Background())

	// Assert
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), now, 2*time.Second)
}
