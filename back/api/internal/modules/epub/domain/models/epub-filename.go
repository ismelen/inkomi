package models

import (
	"path/filepath"
	"strings"
)

const KepubExt = ".kepub.epub"

// CheckEpubFilename accepts plain .epub files, the only input of this worker.
func CheckEpubFilename(filename string) error {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, KepubExt):
		return ErrKepubNotAccepted
	case filepath.Ext(lower) != ".epub":
		return ErrNotEpub
	default:
		return nil
	}
}

// KepubFilename turns "book.epub" into "book.kepub.epub".
func KepubFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + KepubExt
}
