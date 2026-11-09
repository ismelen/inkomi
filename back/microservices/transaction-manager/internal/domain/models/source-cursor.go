package models

import (
	"encoding/base64"
	"strings"
	"time"
)

// SourceCursor points to the last root of a page, ordered by createdAt desc, id desc.
type SourceCursor struct {
	CreatedAt time.Time
	Id        string
}

func NewSourceCursor(source *Source) *SourceCursor {
	return &SourceCursor{CreatedAt: source.CreatedAt, Id: source.Id}
}

func DecodeSourceCursor(raw string) (*SourceCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	createdAt, id, ok := strings.Cut(string(data), "|")
	if !ok || id == "" {
		return nil, ErrInvalidCursor
	}

	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return &SourceCursor{CreatedAt: t.UTC(), Id: id}, nil
}

func (c *SourceCursor) Encode() string {
	raw := c.CreatedAt.UTC().Format(time.RFC3339) + "|" + c.Id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
