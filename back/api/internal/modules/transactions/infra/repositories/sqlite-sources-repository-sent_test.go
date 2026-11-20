package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/infra/repositories"
)

func TestSQLiteSourceRepository_ListSentBooks_MixedSources_ReturnsSentEpubRootsOldestFirst(t *testing.T) {
	t.Parallel()

	// Arrange
	db := setupTestDB(t)
	repo := repositories.NewSQLiteSourceRepository(db)
	file := func(userId int, title, filename string, status models.SourceStatus, minutes int) string {
		id := createAt(t, db, repo, &models.Source{UserId: userId, Title: title, Filename: filename, IsFile: true}, listBase.Add(time.Duration(minutes)*time.Minute))
		setStatus(t, db, id, status)
		return id
	}
	newer := file(1, "Las madres", "Las madres.epub", models.SourceStatusSent, 20)
	older := file(2, "La nena", "La nena.EPUB", models.SourceStatusDone, 10)
	file(1, "Chapter 1", "Chapter 1.cbz", models.SourceStatusSent, 15)
	file(1, "Pending", "Pending.epub", models.SourceStatusProcessing, 15)
	file(1, "Before", "Before.epub", models.SourceStatusSent, -10)
	folder := createAt(t, db, repo, &models.Source{UserId: 1, Title: "Folder", Filename: "Folder"}, listBase.Add(15*time.Minute))
	setStatus(t, db, folder, models.SourceStatusSent)

	// Act
	books, err := repo.ListSentBooks(context.Background(), listBase, 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, books, 2)
	assert.Equal(t, older, books[0].SourceId)
	assert.Equal(t, 2, books[0].UserId)
	assert.Equal(t, "La nena", books[0].Title)
	assert.Equal(t, listBase.Add(10*time.Minute), books[0].UpdatedAt)
	assert.Equal(t, newer, books[1].SourceId)
}
