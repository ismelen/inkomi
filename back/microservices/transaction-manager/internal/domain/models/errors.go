package models

import "errors"

var (
	ErrSourceNotFound    = errors.New("source not found")
	ErrClientNotConnected = errors.New("client not connected")
	ErrSourceNotUploaded = errors.New("source not uploaded")
	ErrSourceNotReady    = errors.New("source not ready")
	ErrFileNotAvailable  = errors.New("file not available")
	ErrInvalidCursor     = errors.New("invalid cursor")
	ErrInvalidLimit      = errors.New("invalid limit")
)
