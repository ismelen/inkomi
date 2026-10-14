package repositories_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLiteSourceRepository_Create_ValidSource_ReturnsId(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)

	source := &models.Source{
		UserID:   1,
		Filename: "test.zip",
		Title:    "Test Title",
		Join:     true,
		IsRTL:    true,
		IsFile:   true,
	}

	// Act
	id, err := repo.Create(context.Background(), source)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM source WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLiteSourceRepository_GetByIDAndUserID_ExistingSource_ReturnsSource(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	size := int64(42)
	created := &models.Source{
		UserID:   1,
		Title:    "Title",
		Filename: "book.epub",
		Kepubify: true,
		ToCloud:  true,
		IsRTL:    true,
		Size:     &size,
		IsFile:   true,
	}
	id, err := repo.Create(context.Background(), created)
	require.NoError(t, err)

	// Act
	source, err := repo.GetByIDAndUserID(context.Background(), id, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, id, source.ID)
	assert.Equal(t, 1, source.UserID)
	assert.Equal(t, "Title", source.Title)
	assert.Equal(t, "book.epub", source.Filename)
	assert.True(t, source.Kepubify)
	assert.True(t, source.ToCloud)
	assert.True(t, source.IsRTL)
	assert.True(t, source.IsFile)
	assert.False(t, source.Join)
	assert.Equal(t, models.SourceStatusPendingUpload, source.Status)
	assert.Equal(t, &size, source.Size)
	assert.Nil(t, source.ParentID)
	assert.Nil(t, source.ConfigKey)
	assert.False(t, source.CreatedAt.IsZero())
}

func TestSQLiteSourceRepository_GetByIDAndUserID_OtherUser_ReturnsNotFound(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	id, err := repo.Create(context.Background(), &models.Source{UserID: 1, Filename: "a.epub"})
	require.NoError(t, err)

	// Act
	source, err := repo.GetByIDAndUserID(context.Background(), id, 2)

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotFound)
	assert.Nil(t, source)
}

func TestSQLiteSourceRepository_GetDetailByIDAndUserID_WithConfigAndParent_LoadsRelations(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	configRepo := repositories.NewSQLiteConfigRepository(db)
	repo := repositories.NewSQLiteSourceRepository(db)

	hash, err := configRepo.Create(ctx, &models.Config{Hash: "cfg-hash", Data: `{"gamma":1}`})
	require.NoError(t, err)

	parentID, err := repo.Create(ctx, &models.Source{UserID: 1, Title: "Folder", IsRTL: true, Join: true})
	require.NoError(t, err)

	childID, err := repo.Create(ctx, &models.Source{
		UserID:    1,
		Title:     "Chapter 1",
		Filename:  "ch1.cbz",
		IsFile:    true,
		ConfigKey: &hash,
		ParentID:  &parentID,
	})
	require.NoError(t, err)

	// Act
	detail, err := repo.GetDetailByIDAndUserID(ctx, childID, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, childID, detail.ID)
	assert.Equal(t, &parentID, detail.ParentID)

	require.NotNil(t, detail.Config)
	assert.Equal(t, "cfg-hash", detail.Config.Hash)
	assert.Equal(t, `{"gamma":1}`, detail.Config.Data)

	require.NotNil(t, detail.Parent)
	assert.Equal(t, models.SourceParent{ID: parentID, Title: "Folder", IsRTL: true, Join: true}, *detail.Parent)
}

func TestSQLiteSourceRepository_GetDetailByIDAndUserID_WithoutRelations_ReturnsNilRelations(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	id, err := repo.Create(context.Background(), &models.Source{UserID: 1, Filename: "book.epub", IsFile: true})
	require.NoError(t, err)

	// Act
	detail, err := repo.GetDetailByIDAndUserID(context.Background(), id, 1)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, detail.Config)
	assert.Nil(t, detail.Parent)
}

func TestSQLiteSourceRepository_GetDetailByIDAndUserID_NotExisting_ReturnsNotFound(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)

	// Act
	detail, err := repo.GetDetailByIDAndUserID(context.Background(), "missing", 1)

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotFound)
	assert.Nil(t, detail)
}

func TestSQLiteSourceRepository_UpdateStatus_ExistingSource_UpdatesStatus(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	source := &models.Source{UserID: 1, Filename: "test.zip"}
	id, _ := repo.Create(context.Background(), source)

	// Act
	err := repo.UpdateStatus(context.Background(), id, 1, models.SourceStatusFailed)

	// Assert
	assert.NoError(t, err)

	var status string
	err = db.QueryRow("SELECT status FROM source WHERE id = ?", id).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, string(models.SourceStatusFailed), status)
}

func TestSQLiteSourceRepository_Delete_ExistingSource_DeletesRecord(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	source := &models.Source{UserID: 1, Filename: "test.zip"}
	id, _ := repo.Create(context.Background(), source)

	// Act
	err := repo.Delete(context.Background(), id, 1)

	// Assert
	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM source WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
