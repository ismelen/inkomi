package manga

import (
	"image"
)

type PagePayload struct {
	Mode            string // N, R, S1, S2
	Image           *image.Image
	OriginalMode    string
	TargetPathOrder string
	Rotated         bool
	Path            string
	Title           string
	Ext             string
	H int
	W int
}

func NewPagePayload(mode string, img *image.Image) *PagePayload {
	originalMode := "RGB"
	if _, ok := (*img).(*image.Gray); ok {
		originalMode = "L"
	} else if _, ok := (*img).(*image.Gray16); ok {
		originalMode = "L"
	}

	payload := &PagePayload{
		OriginalMode: originalMode,
		Image:        img,
		Rotated:      false,
	}

	switch mode {
	case "N":
		payload.TargetPathOrder = "-kcc-x"
		break
	case "R":
		payload.Rotated = true
		payload.TargetPathOrder = "-kcc-d"
		break
	case "S1":
		payload.TargetPathOrder = "-kcc-b"
		break
	case "S2":
		payload.TargetPathOrder = "-kcc-c"
		break
	}

	return payload
}
