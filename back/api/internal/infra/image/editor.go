package image

import (
	"image"
	"ismelen/inkomi/internal/domain/manga"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type SplitOperation = manga.SplitOperation

type ImageEditor struct {
	Img              *image.Image
	targetW, targetH int
	forceColor       bool
	hasWhiteBg       bool
	SplitOperation   SplitOperation
}

func NewEditor(path string, targetW, targetH int, forceColor bool) (*ImageEditor, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}

	return &ImageEditor{
		Img:            &img,
		targetW:        targetW,
		targetH:        targetH,
		forceColor:     forceColor,
		SplitOperation: manga.SplitNone,
	}, nil
}

func (ip ImageEditor) Copy(img *image.Image, splitOperation SplitOperation) *ImageEditor {
	ip.Img = img
	ip.SplitOperation = splitOperation
	return &ip
}
