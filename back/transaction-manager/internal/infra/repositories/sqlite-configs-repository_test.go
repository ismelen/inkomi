package repositories_test

import (
	"context"
	
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/assert"
)



func TestSQLiteConfigRepository_Create_NewConfig_ReturnsHash(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteConfigRepository(db)

	config := &models.Config{
		Hash: "test-hash",
		Data: `{"key": "value"}`,
	}

	// Act
	hash, err := repo.Create(context.Background(), config)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-hash", hash)

	var data string
	err = db.QueryRow("SELECT data FROM configs WHERE hash = ?", hash).Scan(&data)
	assert.NoError(t, err)
	assert.Equal(t, `{"key": "value"}`, data)
}

func TestSQLiteConfigRepository_Create_ExistingConfig_UpdatesLastUsed(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteConfigRepository(db)

	config := &models.Config{
		Hash: "conflict-hash",
		Data: `{"key": "value"}`,
	}

	_, err := repo.Create(context.Background(), config)
	assert.NoError(t, err)

	var lastUsed1 string
	db.QueryRow("SELECT lastUsed FROM configs WHERE hash = ?", config.Hash).Scan(&lastUsed1)

	time.Sleep(1 * time.Second)

	// Act
	_, err = repo.Create(context.Background(), config)
	
	// Assert
	assert.NoError(t, err)

	var lastUsed2 string
	db.QueryRow("SELECT lastUsed FROM configs WHERE hash = ?", config.Hash).Scan(&lastUsed2)

	assert.NotEqual(t, lastUsed1, lastUsed2, "expected lastUsed to be updated")
}

func TestSQLiteConfigRepository_GetByHash_ExistingConfig_ReturnsConfig(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteConfigRepository(db)

	config := &models.Config{
		Hash: "get-hash",
		Data: `{"get": "data"}`,
	}
	_, _ = repo.Create(context.Background(), config)

	// Act
	res, err := repo.GetByHash(context.Background(), "get-hash")
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)
	if res != nil {
		assert.Equal(t, "get-hash", res.Hash)
		assert.Equal(t, `{"get": "data"}`, res.Data)
	}
}

func TestSQLiteConfigRepository_GetByHash_NotExistingConfig_ReturnsNil(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewSQLiteConfigRepository(db)

	// Act
	res, err := repo.GetByHash(context.Background(), "non-existent")
	
	// Assert
	assert.NoError(t, err)
	assert.Nil(t, res)
}
