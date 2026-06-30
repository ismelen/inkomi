package manga

import (
	"image"
	"ismelen/inkomi/internal/shared/strutil"
	"math"
	"path/filepath"
	"strings"
)

type Page struct {
	HasWhiteBg bool
	Path       string
	Parts      []*PagePart
}

func NewPage(path string) *Page { return &Page{Path: path} }

func (this *Page) GetCSSBgStyle() string {
	if !this.HasWhiteBg {
		return ""
	}
	return "background-color:#000000;"
}

type PagePart struct {
	Img           *image.Image
	PathOrder     rune // X, D, B, C
	Width, Height int
	Path          string
	Name          string
	Ext           string
	ChapterName   string
	Split         SplitOperation
}

func NewPagePart(img *image.Image, splitOperation SplitOperation) *PagePart {
	var pathOrder rune
	switch splitOperation {
	case SplitNone:
		pathOrder = 'X'
	case SplitRotated:
		pathOrder = 'D'
	case SplitToLeft:
		pathOrder = 'B'
	case SplitToRight:
		pathOrder = 'C'
	}

	dim := (*img).Bounds()
	return &PagePart{
		Img:       img,
		Width:     dim.Dx(),
		Height:    dim.Dy(),
		PathOrder: pathOrder,
		Split:     splitOperation,
	}
}

func (pp *PagePart) SetPath(path string) {
	chapterName := filepath.Base(filepath.Dir(path))
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	pp.Path = path
	pp.Name = name
	pp.Ext = ext
	pp.ChapterName = strutil.NormalizeString(chapterName)
}

func (pp *PagePart) GetTopMargin(deviceHeight int) float64 {
	y := ((deviceHeight - pp.Height) / 2) / deviceHeight * 100
	return math.Round(float64(y*10)) / 10
}

func (pp *PagePart) Clean() {
	pp.Img = nil
}
