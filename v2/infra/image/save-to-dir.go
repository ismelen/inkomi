package image

import (
	"ismelen/ermc/internal/pkg"
)

func (ip *ImageEditor) SaveToDir(path string) (string, error) {
	return pkg.SaveWithCodec(
		ip.Img,
		path,
	)
}
