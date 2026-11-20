package models

import "errors"

var (
	ErrInvalidSection = errors.New("invalid section")
	ErrEmptyWorkKey   = errors.New("empty work key")
	ErrNoModel        = errors.New("embedding model unavailable")
)
