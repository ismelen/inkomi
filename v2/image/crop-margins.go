package image

import (
	"image"

	"github.com/disintegration/imaging"
)

func (ip *ImageProcessor) CropMargins() {
	img := image.Image(imaging.AdjustContrast(*ip.Img, 100))
	img = imaging.Grayscale(img)

	box := ip.GetBBox(img, ip.hasWhiteBg)

	rect := image.Rect(
		box.left,
		box.top,
		box.right,
		box.bottom,
	)

	(*ip.Img) = imaging.Crop(*ip.Img, rect)
}
