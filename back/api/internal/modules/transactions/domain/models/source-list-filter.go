package models

import "time"

const (
	DefaultSourceListLimit = 50
	MaxSourceListLimit     = 100
)

type SourceListFilter struct {
	Limit        int
	Cursor       *SourceCursor
	UpdatedSince *time.Time
}

// NewSourceListFilter validates the limit and decodes the cursor; an empty cursor starts from the newest root.
func NewSourceListFilter(limit int, cursor string, updatedSince *time.Time) (*SourceListFilter, error) {
	if limit < 1 || limit > MaxSourceListLimit {
		return nil, ErrInvalidLimit
	}

	filter := &SourceListFilter{Limit: limit, UpdatedSince: updatedSince}
	if cursor != "" {
		c, err := DecodeSourceCursor(cursor)
		if err != nil {
			return nil, err
		}
		filter.Cursor = c
	}

	return filter, nil
}
