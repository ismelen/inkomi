package repositories_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteSourceRepository_UpdateFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		filename         string
		size             int64
		expectedFilename string
		expectedSize     int64
	}{
		{"Both", "book.kepub.epub", 200, "book.kepub.epub", 200},
		{"Only filename", "book.kepub.epub", 0, "book.kepub.epub", 100},
		{"Only size", "", 200, "book.epub", 200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			db := setupTestDB(t)
			defer db.Close()

			ctx := context.Background()
			repo := repositories.NewSQLiteSourceRepository(db)
			size := int64(100)
			id, err := repo.Create(ctx, &models.Source{UserId: 1, Title: "Book", Filename: "book.epub", Size: &size, IsFile: true})
			require.NoError(t, err)

			// Act
			err = repo.UpdateFile(ctx, id, 1, tc.filename, tc.size)

			// Assert
			require.NoError(t, err)
			got, err := repo.GetByIdAndUserId(ctx, id, 1)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedFilename, got.Filename)
			assert.Equal(t, tc.expectedSize, *got.Size)
		})
	}
}

func TestSQLiteSourceRepository_UpsertResult_EmptyFilename_KeepsStoredFilename(t *testing.T) {
	t.Parallel()

	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repositories.NewSQLiteSourceRepository(db)
	parentId := createFolder(t, repo)
	result := resultSource(parentId, 0, models.SourceStatusKepubifying)
	require.NoError(t, repo.UpsertResult(ctx, result))

	// Act
	result.Filename = ""
	result.Status = models.SourceStatusFailed
	err := repo.UpsertResult(ctx, result)

	// Assert
	require.NoError(t, err)
	got, err := repo.GetByIdAndUserId(ctx, result.Id, 1)
	require.NoError(t, err)
	assert.Equal(t, "Folder Ch.[1 - 2].epub", got.Filename)
}
