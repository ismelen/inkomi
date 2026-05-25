package domain

type Chapter struct {
	Path      string
	Filename  string
	Size      int64
	PagePaths []string
}

func NewChapter(filename, path string, pagePahts []string, size int64) *Chapter {
	return &Chapter{
		Path:      path,
		PagePaths: pagePahts,
		Size:      size,
		Filename:  filename,
	}
}