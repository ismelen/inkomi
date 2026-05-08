package domain

type Chapter struct {
	Path      string
	Size      int64
	PagePaths []string
}

func NewChapter(path string, pagePahts []string, size int64) *Chapter {
	return &Chapter{
		Path:      path,
		PagePaths: pagePahts,
		Size:      size,
	}
}