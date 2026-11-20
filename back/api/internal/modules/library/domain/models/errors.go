package models

import "errors"

var (
	ErrInvalidMD5      = errors.New("invalid md5")
	ErrEmptyQuery      = errors.New("empty query")
	ErrBookNotFound    = errors.New("book not found")
	ErrNoMirror        = errors.New("no libgen mirror available")
	ErrNoResults       = errors.New("no books found")
	ErrInvalidField    = errors.New("invalid search field")
	ErrInvalidCoverURL = errors.New("invalid cover url")
)
