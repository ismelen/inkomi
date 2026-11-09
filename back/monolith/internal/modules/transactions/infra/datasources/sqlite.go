package datasources

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLiteDatasource struct {
	DB *sql.DB
}

// NewSQLiteDatasource opens the database at dbPath and applies schema to create the tables.
func NewSQLiteDatasource(dbPath, schema string) (*SQLiteDatasource, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to ping database: %w", err), db.Close())
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to execute schema: %w", err), db.Close())
	}

	return &SQLiteDatasource{DB: db}, nil
}

func (ds *SQLiteDatasource) Close() error {
	return ds.DB.Close()
}

// --- Transaction Support ---

type IDBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type txKey struct{}

// ExecuteTx executes a function within a database transaction.
// If the function returns an error, it rolls back. Otherwise, it commits.
func (ds *SQLiteDatasource) ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := ds.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetDB returns the transaction from the context if it exists,
// otherwise it returns the default DB connection.
func GetDB(ctx context.Context, defaultDB *sql.DB) IDBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultDB
}
