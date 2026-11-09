package repositories_test

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

const migrationPath = "../../../migrations/transactions.sql"

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// Every :memory: connection is a different database.
	db.SetMaxOpenConns(1)

	schema, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}
