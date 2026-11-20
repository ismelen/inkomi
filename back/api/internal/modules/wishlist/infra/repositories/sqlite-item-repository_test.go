package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
	"github.com/ismelen/inkomi/back/api/migrations"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(migrations.Wishlist)
	require.NoError(t, err)
	return db
}

func item(t *testing.T, userId int, kind models.ItemKind, title, author string) *models.Item {
	t.Helper()

	it, err := models.NewItem(userId, kind, title, author, "", "")
	require.NoError(t, err)
	return it
}

func TestSQLiteItemRepository_AddListDelete_Lifecycle_Works(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteItemRepository(setupDB(t))
	first := item(t, 1, models.ItemKindBook, "La nena", "Carmen Mola")
	second := item(t, 1, models.ItemKindAuthor, "Frank Herbert", "")
	other := item(t, 2, models.ItemKindBook, "Dune", "")

	// Act
	require.NoError(t, repo.Add(ctx, first))
	require.NoError(t, repo.Add(ctx, second))
	require.NoError(t, repo.Add(ctx, other))
	duplicate := repo.Add(ctx, item(t, 1, models.ItemKindBook, "LA NENA", "carmen mola"))
	mine, err := repo.ListByUserId(ctx, 1)
	require.NoError(t, err)
	count, err := repo.CountByUserId(ctx, 1)
	require.NoError(t, err)
	notMine := repo.Delete(ctx, other.Id, 1)
	deleted := repo.Delete(ctx, first.Id, 1)
	afterDelete, err := repo.ListByUserId(ctx, 1)
	require.NoError(t, err)

	// Assert
	assert.ErrorIs(t, duplicate, models.ErrDuplicate)
	require.Len(t, mine, 2)
	assert.Equal(t, second.Id, mine[0].Id, "newest first")
	assert.Equal(t, "Carmen Mola", mine[1].Author)
	assert.False(t, mine[1].CreatedAt.IsZero())
	assert.Equal(t, 2, count)
	assert.ErrorIs(t, notMine, models.ErrItemNotFound)
	require.NoError(t, deleted)
	assert.Len(t, afterDelete, 1)
}
