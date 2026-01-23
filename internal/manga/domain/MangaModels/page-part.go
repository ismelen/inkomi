package MangaModels

import (
	"image"
)

type PagePart struct {
	Mode            rune // N, R, 1, 2
	Rotated         bool
	H               int
	W               int
	Image           *image.Image
	TargetPathOrder string
	OriginalMode    string
	Path            string
	Title           string
	Ext             string
}

var PathOrders = map[rune]string{
	'N': "-ermc-x",
	'R': "-ermc-d",
	'1': "-ermc-b",
	'2': "-ermc-c",
}

func NewPagePart(mode rune, img *image.Image) *PagePart {
	originalMode := "RGB"
	if _, ok := (*img).(*image.Gray); ok {
		originalMode = "L"
	} else if _, ok := (*img).(*image.Gray16); ok {
		originalMode = "L"
	}

	payload := &PagePart{
		OriginalMode: originalMode,
		Image:        img,
		Rotated:      false,
	}

	if value, ok := PathOrders[mode]; ok {
		payload.TargetPathOrder = value
	} else {
		payload.TargetPathOrder = "-ermc-x"
	}

	payload.Rotated = mode == 'R'

	return payload
}
