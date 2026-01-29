package image

import (
	"image"

	"github.com/disintegration/imaging"
)

func (ip *ImageProcessor) TrySplit(rotated bool) []*ImageProcessor {
	dim := (*ip.Img).Bounds()
	w, h := dim.Dx(), dim.Dy()

	if (w > h) == (ip.targetW > ip.targetH) {
		return []*ImageProcessor{ip}
	}

	if w <= ip.targetH &&
		h <= ip.targetW &&
		rotated {
		spread := image.Image(imaging.Rotate270(*ip.Img))
		return []*ImageProcessor{ip.Copy(&spread, Rotated)}
	}

	var processors []*ImageProcessor

	if !rotated {
		var leftBox, rightBox image.Rectangle
		if w < h {
			leftBox = image.Rect(0, 0, w, h/2)
			rightBox = image.Rect(0, w/2, w, h)
		} else {
			leftBox = image.Rect(0, 0, w/2, h)
			rightBox = image.Rect(w/2, 0, w, h)
		}

		var pageOne image.Image = imaging.Crop(*ip.Img, leftBox)
		var pageTwo image.Image = imaging.Crop(*ip.Img, rightBox)
		processors = append(processors,
			ip.Copy(&pageOne, ToLeft),
			ip.Copy(&pageTwo, ToRight),
		)
	}

	spread := image.Image(imaging.Rotate270(*ip.Img))
	processors = append(processors, ip.Copy(&spread, Rotated))

	return processors
}
