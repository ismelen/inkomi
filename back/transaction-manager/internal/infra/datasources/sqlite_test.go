package datasources_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteDatasource_ExecuteTx_Success_CommitsTransaction(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test_tx (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	ds := &datasources.SQLiteDatasource{DB: db}

	// Act
	err = ds.ExecuteTx(context.Background(), func(ctx context.Context) error {
		txDB := datasources.GetDB(ctx, db)
		_, err := txDB.ExecContext(ctx, "INSERT INTO test_tx (id) VALUES (1)")
		return err
	})

	// Assert
	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLiteDatasource_ExecuteTx_Error_RollsBackTransaction(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test_tx (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	ds := &datasources.SQLiteDatasource{DB: db}

	expectedErr := errors.New("something went wrong")

	// Act
	err = ds.ExecuteTx(context.Background(), func(ctx context.Context) error {
		txDB := datasources.GetDB(ctx, db)
		_, err := txDB.ExecContext(ctx, "INSERT INTO test_tx (id) VALUES (2)")
		if err != nil {
			return err
		}
		return expectedErr
	})

	// Assert
	assert.EqualError(t, err, "something went wrong")

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
