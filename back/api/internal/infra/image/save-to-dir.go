package image

import (
	"image"

	"github.com/disintegration/imaging"
)

func (ip *ImageEditor) SaveToDir(path string) (string, error) {
	return SaveWithCodec(
		ip.Img,
		path,
	)
}

func SaveWithCodec(img *image.Image, targetPath string) (string, error) {
	targetPath += ".jpg"
	err := imaging.Save((*img), targetPath, imaging.JPEGQuality(85))
	if err != nil {
		return "", err
	}
	return targetPath, nil
}
