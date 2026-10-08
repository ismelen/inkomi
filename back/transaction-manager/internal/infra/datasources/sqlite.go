package datasources

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type SQLiteDatasource struct {
	DB *sql.DB
}

// NewSQLiteDatasource creates a new SQLite connection.
// If schemaPath is provided, it will execute the schema to create tables.
func NewSQLiteDatasource(dbPath, schemaPath string) (*SQLiteDatasource, error) {
	if dbPath == "" {
		dbPath = "transactions.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if schemaPath != "" {
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file: %w", err)
		}

		if _, err := db.Exec(string(schema)); err != nil {
			return nil, fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return &SQLiteDatasource{DB: db}, nil
}

func (ds *SQLiteDatasource) Close() error {
	return ds.DB.Close()
}

// --- Transaction Support ---

type DBTX interface {
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
func GetDB(ctx context.Context, defaultDB *sql.DB) DBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultDB
}
