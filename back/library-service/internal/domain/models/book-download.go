package models

import (
	"io"
	"strings"

	"github.com/ismelen/inkomi/back/common/fsutil"
)

const defaultExt = "epub"

var pathSeparators = strings.NewReplacer("/", "-", "\\", "-")

// BookDownload is an open download of a book file; Stream must be closed by the caller.
type BookDownload struct {
	Stream io.ReadCloser
	Title  string
	Ext    string
}

// Filename returns "<title>.<ext>" safe to use as a file and cloud object name.
func (b *BookDownload) Filename() (string, error) {
	ext := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(b.Ext), "."))
	if ext == "" {
		ext = defaultExt
	}
	return fsutil.SanitizeFilename(pathSeparators.Replace(strings.TrimSpace(b.Title)) + "." + ext)
}
