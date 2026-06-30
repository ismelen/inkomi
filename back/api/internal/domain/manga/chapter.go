package manga

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

func NewChapter(filename, path string, pagePaths []string) *Chapter {
	return &Chapter{
		Path:      path,
		PagePaths: pagePaths,
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

	var validPaths []string
	for _, path := range c.PagePaths {
		if filepath.Ext(path) == ".xml" {
			continue
		}
		validPaths = append(validPaths, path)
	}

	c.PagePaths = validPaths
	c.ordered = true
	return c.PagePaths
}
