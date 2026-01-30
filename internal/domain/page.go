package domain

import (
	"image"
	IMG "ismelen/ermc/internal/image"
	"ismelen/ermc/internal/pkg"
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
	Ext string
	ChapterName   string
	Split         IMG.SplitOperation
}

func NewPagePart(img *image.Image, splitOperation IMG.SplitOperation) *PagePart {
	var pathOrder rune
	switch splitOperation {
	case IMG.None:
		pathOrder = 'X'
	case IMG.Rotated:
		pathOrder = 'D'
	case IMG.ToLeft:
		pathOrder = 'B'
	case IMG.ToRight:
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
	pp.ChapterName = pkg.NormalizeString(chapterName)
}

func (pp *PagePart) GetTopMargin(deviceHeight int) float64 {
	y := ((deviceHeight - pp.Height) / 2) / deviceHeight * 100
	return math.Round(float64(y*10)) / 10
}

func (pp *PagePart) Clean() {
	pp.Img = nil
}
