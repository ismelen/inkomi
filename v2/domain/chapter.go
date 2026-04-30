package domain

type Chapter struct {
	Path      string
	PagePaths []string
}

func NewChapter(path string, pagePahts []string) *Chapter {
	return &Chapter{
		Path:      path,
		PagePaths: pagePahts,
	}
}