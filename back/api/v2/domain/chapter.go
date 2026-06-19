package domain

import "sort"

type Chapter struct {
	Path      string
	Filename  string
	Size      int64
	pagePaths []string
}

func NewChapter(filename, path string, pagePahts []string, size int64) *Chapter {
	return &Chapter{
		Path:      path,
		pagePaths: pagePahts,
		Size:      size,
		Filename:  filename,
	}
}

func (c *Chapter) GetPagePaths() []string {
	sort.Slice(c.pagePaths, func(i, j int) bool {
		return c.pagePaths[i] < c.pagePaths[j]
	})
	return c.pagePaths
}
