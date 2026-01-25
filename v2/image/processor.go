package image

import (
	"image"

	"github.com/disintegration/imaging"
)

type SplitOperation = int

const (
	None SplitOperation = iota
	Rotated
	ToRight
	ToLeft
)

type ImageProcessor struct {
	Img              *image.Image
	targetW, targetH int
	forceColor       bool
	hasWhiteBg       bool
	SplitOperation   SplitOperation
}

func NewProcessor(path string, targetW, targetH int, forceColor bool) (*ImageProcessor, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}

	return &ImageProcessor{
		Img:            &img,
		targetW:        targetW,
		targetH:        targetH,
		forceColor:     forceColor,
		SplitOperation: None,
	}, nil
}

func (ip ImageProcessor) Copy(img *image.Image, splitOperation SplitOperation) *ImageProcessor {
	ip.Img = img
	ip.SplitOperation = splitOperation
	return &ip
}
