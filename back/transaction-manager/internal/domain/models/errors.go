package models

import "errors"

var (
	ErrSourceNotFound    = errors.New("source not found")
	ErrSourceNotUploaded = errors.New("source not uploaded")
)
