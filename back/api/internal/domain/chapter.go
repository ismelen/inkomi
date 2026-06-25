package domain

import (
	"path/filepath"
	"sort"
)

type Chapter struct {
	Path      string
	Filename  string
	PagePaths []string
	ordered   bool
}

func NewChapter(filename, path string, pagePahts []string) *Chapter {
	return &Chapter{
		Path:      path,
		PagePaths: pagePahts,
		Filename:  filename,
		ordered:   false,
	}
}

func (c *Chapter) GetOrderedPagePaths() []string {
	if c.ordered {
		return c.PagePaths
	}

	sort.Slice(c.PagePaths, func(i, j int) bool {
		return c.PagePaths[i] < c.PagePaths[j]
	})

	var validPahts []string
	for _, path := range c.PagePaths {
		if filepath.Ext(path) == ".xml" {
			continue
		}
		validPahts = append(validPahts, path)
	}

	c.PagePaths = validPahts
	c.ordered = true
	return c.PagePaths
}
