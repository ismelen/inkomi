package repositories_test

import (
	"database/sql"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/migrations"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// Every :memory: connection is a different database.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(migrations.Transactions); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}
