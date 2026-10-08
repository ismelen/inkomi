package repositories_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE configs (
		hash TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		lastUsed DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE sources (
		id TEXT PRIMARY KEY,
		userId INTEGER NOT NULL,
		size INTEGER,
		filename TEXT,
		title TEXT,
		kepubify BOOLEAN,
		type TEXT,
		status TEXT,
		error TEXT,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		completedAt DATETIME,
		shouldJoin BOOLEAN,
		readingDirection TEXT,
		folderId TEXT,
		configHash TEXT,
		FOREIGN KEY(configHash) REFERENCES configs(hash)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

