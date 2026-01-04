package manga

import (
	"image"
)

type PageData struct {
	Src      string
	Img      *image.Image
	Payloads []*PagePayload
	Fill     string
}

func NewPageData(path string) *PageData {
	return &PageData{Src: path}
}
