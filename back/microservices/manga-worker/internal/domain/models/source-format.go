package models

import (
	"path/filepath"
	"strings"
)

type SourceFormat string

const (
	SourceFormatCBZ SourceFormat = "cbz"
	SourceFormatPDF SourceFormat = "pdf"
)

func SourceFormatFromFilename(filename string) (SourceFormat, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".cbz", ".zip":
		return SourceFormatCBZ, nil
	case ".pdf":
		return SourceFormatPDF, nil
	default:
		return "", ErrUnsupportedFormat
	}
}
