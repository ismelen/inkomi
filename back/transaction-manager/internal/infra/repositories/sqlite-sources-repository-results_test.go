package repositories_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createFolder(t *testing.T, repo ports.SourceRepository) string {
	t.Helper()

	id, err := repo.Create(context.Background(), &models.Source{UserId: 1, Title: "Folder", Join: true})
	require.NoError(t, err)
	return id
}

func resultSource(parentId string, position int, status models.SourceStatus) *models.Source {
	size := int64(100)
	return &models.Source{
		Id:       fmt.Sprintf("result-%d", position),
		UserId:   1,
		Title:    "Folder Ch.[1 - 2]",
		Filename: "Folder Ch.[1 - 2].epub",
		Status:   status,
		Size:     &size,
		ParentId: &parentId,
		Position: position,
	}
}

func TestSQLiteSourceRepository_UpsertResult_New_CreatesResult(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	parentId := createFolder(t, repo)
	result := resultSource(parentId, 1, models.SourceStatusKepubifying)

	// Act
	err := repo.UpsertResult(context.Background(), result)

	// Assert
	require.NoError(t, err)
	got, err := repo.GetByIdAndUserId(context.Background(), result.Id, 1)
	require.NoError(t, err)
	assert.True(t, got.IsResult)
	assert.True(t, got.IsFile)
	assert.Equal(t, models.SourceStatusKepubifying, got.Status)
	assert.Equal(t, &parentId, got.ParentId)
	assert.Equal(t, 1, got.Position)
	assert.Equal(t, int64(100), *got.Size)
	assert.Nil(t, got.CompletedAt)
}

func TestSQLiteSourceRepository_UpsertResult_Existing_UpdatesStatusAndKeepsCreatedAt(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	parentId := createFolder(t, repo)
	result := resultSource(parentId, 0, models.SourceStatusKepubifying)
	require.NoError(t, repo.UpsertResult(ctx, result))
	_, err := db.Exec(`UPDATE source SET createdAt = '2000-01-01 00:00:00' WHERE id = ?`, result.Id)
	require.NoError(t, err)

	// Act
	result.Status = models.SourceStatusDone
	err = repo.UpsertResult(ctx, result)

	// Assert
	require.NoError(t, err)
	got, err := repo.GetByIdAndUserId(ctx, result.Id, 1)
	require.NoError(t, err)
	assert.Equal(t, models.SourceStatusDone, got.Status)
	assert.Equal(t, 2000, got.CreatedAt.Year())
	assert.NotNil(t, got.CompletedAt)

	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM source WHERE isResult = TRUE`).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestSQLiteSourceRepository_ListResultsByParentIdAndUserId_ReturnsOnlyResultsInOrder(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	parentId := createFolder(t, repo)
	_, err := repo.Create(ctx, &models.Source{UserId: 1, Title: "Child", Filename: "ch.epub", IsFile: true, ParentId: &parentId})
	require.NoError(t, err)
	require.NoError(t, repo.UpsertResult(ctx, resultSource(parentId, 1, models.SourceStatusDone)))
	require.NoError(t, repo.UpsertResult(ctx, resultSource(parentId, 0, models.SourceStatusDone)))

	// Act
	results, err := repo.ListResultsByParentIdAndUserId(ctx, parentId, 1)

	// Assert
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 0, results[0].Position)
	assert.Equal(t, 1, results[1].Position)
}

func TestSQLiteSourceRepository_GetDetailByIdAndUserId_WithResults_ExcludesResultsFromChildCount(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	parentId := createFolder(t, repo)
	childId, err := repo.Create(ctx, &models.Source{UserId: 1, Title: "Child", Filename: "ch.epub", IsFile: true, ParentId: &parentId, Position: 3})
	require.NoError(t, err)
	require.NoError(t, repo.UpsertResult(ctx, resultSource(parentId, 0, models.SourceStatusDone)))

	// Act
	detail, err := repo.GetDetailByIdAndUserId(ctx, childId, 1)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, detail.Parent)
	assert.Equal(t, 1, detail.Parent.ChildCount)
	assert.Equal(t, 3, detail.Position)
	assert.False(t, detail.IsResult)
}
