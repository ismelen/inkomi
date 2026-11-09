package models

import "errors"

var (
	ErrNotEpub          = errors.New("only .epub files are accepted")
	ErrKepubNotAccepted = errors.New(".kepub.epub files are not accepted")
)
