package repositories_test

import (
	"context"
	
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)



func TestSQLiteSourceRepository_Create_ValidSource_ReturnsId(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)

	source := &models.Source{
		UserId:           1,
		Filename:         "test.zip",
		Title:            "Test Title",
		ShouldJoin:       true,
		ReadingDirection: models.ReadingDirectionRTL,
	}

	// Act
	id, err := repo.Create(context.Background(), source)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLiteSourceRepository_UpdateStatus_ExistingSource_UpdatesStatus(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	source := &models.Source{UserId: 1, Filename: "test.zip"}
	id, _ := repo.Create(context.Background(), source)

	errStr := "some error"

	// Act
	err := repo.UpdateStatus(context.Background(), id, 1, models.SourceStatusFailed, &errStr)

	// Assert
	assert.NoError(t, err)

	var status, queryErr string
	err = db.QueryRow("SELECT status, error FROM sources WHERE id = ?", id).Scan(&status, &queryErr)
	assert.NoError(t, err)
	assert.Equal(t, string(models.SourceStatusFailed), status)
	assert.Equal(t, "some error", queryErr)
}

func TestSQLiteSourceRepository_Delete_ExistingSource_DeletesRecord(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteSourceRepository(db)
	source := &models.Source{UserId: 1, Filename: "test.zip"}
	id, _ := repo.Create(context.Background(), source)

	// Act
	err := repo.Delete(context.Background(), id, 1)

	// Assert
	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
