package models

import "errors"

var (
	ErrInvalidMangaConfig = errors.New("invalid manga config")
	ErrUnsupportedFormat  = errors.New("only .cbz, .zip and .pdf files are accepted")
	ErrNoPages            = errors.New("no image pages found")
)
