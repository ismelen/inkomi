package datasources

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	_ "modernc.org/sqlite"
)

var _ ports.ITxManager = (*SQLiteDatasource)(nil)

type SQLiteDatasource struct {
	DB *sql.DB
}

// NewSQLiteDatasource opens the database at dbPath with foreign keys enabled and applies schema.
func NewSQLiteDatasource(dbPath, schema string) (*SQLiteDatasource, error) {
	dsn := dbPath + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Join(fmt.Errorf("ping database: %w", err), db.Close())
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, errors.Join(fmt.Errorf("apply schema: %w", err), db.Close())
	}
	return &SQLiteDatasource{DB: db}, nil
}

func (s *SQLiteDatasource) Close() error {
	return s.DB.Close()
}

type IDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type txKey struct{}

// ExecuteTx runs fn in a transaction that the repositories pick up from its ctx; it commits
// when fn succeeds and rolls back otherwise.
func (s *SQLiteDatasource) ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return errors.Join(err, ignoreDone(tx.Rollback()))
	}
	return tx.Commit()
}

// GetDB returns the transaction of ctx or, outside of one, db.
func GetDB(ctx context.Context, db *sql.DB) IDBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

func ignoreDone(err error) error {
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}
