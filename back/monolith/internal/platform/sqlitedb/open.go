package sqlitedb

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens the SQLite database at path with WAL and a busy timeout, and applies schema.
func Open(path, schema string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Join(fmt.Errorf("ping database: %w", err), db.Close())
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, errors.Join(fmt.Errorf("apply schema: %w", err), db.Close())
	}
	return db, nil
}
