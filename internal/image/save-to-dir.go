package image

import (
	"ismelen/ermc/internal/pkg"
)

func (ip *ImageProcessor) SaveToDir(path string) (string, error) {
	return pkg.SaveWithCodec(
		ip.Img,
		path,
	)
}
