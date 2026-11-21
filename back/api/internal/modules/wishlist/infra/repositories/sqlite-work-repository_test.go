package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
)

func TestSQLiteWorkRepository_SaveGet_RoundTripsAndReplaces(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteWorkRepository(setupDB(t))
	work := models.WorkDetails{
		WorkKey: "OL5W", Title: "Mistborn", Description: "Ash falls.", CoverURL: "https://covers.openlibrary.org/b/id/1-M.jpg",
		Subjects: []string{"Fantasy", "Magic"}, FirstPublished: "2006",
	}

	// Act
	_, missingErr := repo.Get(ctx, "OL5W")
	require.NoError(t, repo.Save(ctx, work))
	got, err := repo.Get(ctx, "OL5W")
	work.Description = "Updated."
	require.NoError(t, repo.Save(ctx, work))
	updated, updatedErr := repo.Get(ctx, "OL5W")

	// Assert
	assert.ErrorIs(t, missingErr, models.ErrWorkNotFound)
	require.NoError(t, err)
	assert.Equal(t, "Mistborn", got.Title)
	assert.Equal(t, []string{"Fantasy", "Magic"}, got.Subjects)
	require.NoError(t, updatedErr)
	assert.Equal(t, "Updated.", updated.Description)
}

func TestSQLiteItemRepository_Works_JoinsCoverAndListsPendingBooks(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	db := setupDB(t)
	items := repositories.NewSQLiteItemRepository(db)
	works := repositories.NewSQLiteWorkRepository(db)
	typed := item(t, 1, models.ItemKindBook, "Mistborn", "Sanderson")
	author := item(t, 1, models.ItemKindAuthor, "Carmen Mola", "")
	other := item(t, 2, models.ItemKindBook, "Elantris", "")
	for _, it := range []*models.Item{typed, author, other} {
		require.NoError(t, items.Add(ctx, it))
	}

	// Act
	pendingBefore, err := items.ListBooksWithoutWork(ctx)
	require.NoError(t, err)
	require.NoError(t, items.SetWorkKey(ctx, typed, "OL5W"))
	require.NoError(t, works.Save(ctx, models.WorkDetails{WorkKey: "OL5W", CoverURL: "https://covers.openlibrary.org/b/id/1-M.jpg"}))
	pendingAfter, err := items.ListBooksWithoutWork(ctx)
	require.NoError(t, err)
	listed, err := items.ListByUserId(ctx, 1)
	require.NoError(t, err)

	// Assert
	assert.Len(t, pendingBefore, 2, "both books, not the author")
	require.Len(t, pendingAfter, 1)
	assert.Equal(t, other.Id, pendingAfter[0].Id)
	require.Len(t, listed, 2)
	for _, it := range listed {
		if it.Id == typed.Id {
			assert.Equal(t, "OL5W", it.WorkKey)
			assert.Equal(t, "https://covers.openlibrary.org/b/id/1-M.jpg", it.CoverURL)
		} else {
			assert.Empty(t, it.CoverURL)
		}
	}
}

func TestSQLiteItemRepository_SetWorkKey_SameWorkTwice_ReturnsErrDuplicate(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteItemRepository(setupDB(t))
	first := item(t, 1, models.ItemKindBook, "Mistborn", "")
	second := item(t, 1, models.ItemKindBook, "Nacidos de la bruma", "")
	require.NoError(t, repo.Add(ctx, first))
	require.NoError(t, repo.Add(ctx, second))
	require.NoError(t, repo.SetWorkKey(ctx, first, "OL5W"))

	// Act
	err := repo.SetWorkKey(ctx, second, "OL5W")

	// Assert
	assert.ErrorIs(t, err, models.ErrDuplicate)
	assert.Empty(t, second.WorkKey)
}
