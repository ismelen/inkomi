package manga

import (
	"image"
	IMG "ismelen/ermc/v2/image"
	"sync"
)

type Volume struct {
	Name     string
	Chapters []*Chapter
	Wg       *sync.WaitGroup
}

type Chapter struct {
	Path  string
	Name  string
	Size  int // MB
	Pages []*Page
}

type Page struct {
	HasWhiteBg bool
	Path       string
	Parts      []*PagePart
}

type PagePart struct {
	img           *image.Image
	pathOrder     rune // R, L, R
	width, height int
	path          string
	split         IMG.SplitOperation
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
		img:       img,
		width:     dim.Dx(),
		height:    dim.Dy(),
		pathOrder: pathOrder,
		split:     splitOperation,
	}
}
