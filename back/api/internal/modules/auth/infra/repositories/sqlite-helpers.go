package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// checkAffected returns notFound when an update or delete without err matched no rows.
func checkAffected(result sql.Result, err error, notFound error) error {
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return notFound
	}
	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullUnix(t time.Time) sql.NullInt64 {
	return sql.NullInt64{Int64: t.Unix(), Valid: !t.IsZero()}
}

func fromNullUnix(seconds sql.NullInt64) time.Time {
	if !seconds.Valid {
		return time.Time{}
	}
	return time.Unix(seconds.Int64, 0)
}
