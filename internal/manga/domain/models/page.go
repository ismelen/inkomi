package MangaModels

type Page struct {
	Count      int8
	HasWhiteBg bool
	Parts      [3]*PagePart
	Path       string
}

func NewPage(path string) *Page {
	return &Page{Path: path, Count: 1}
}

func (this *Page) GetCSSBgStyle() string {
	if !this.HasWhiteBg {
		return ""
	}

	return "background-color:#000000;"
}
