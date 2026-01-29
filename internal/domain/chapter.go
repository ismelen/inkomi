package domain

import (
	"ismelen/ermc/internal/pkg"
	"path/filepath"
	"strings"
)

type Chapter struct {
	Name  string
	Size  int64 // MB
	Pages []*Page
}

func NewChapter(name string, size int64, pages []*Page) *Chapter {
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)

	return &Chapter{
		Name:  pkg.NormalizeString(name),
		Size:  size,
		Pages: pages,
	}
}
