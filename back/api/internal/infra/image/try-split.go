package image

import (
	"image"
	"ismelen/inkomi/internal/domain/manga"

	"github.com/disintegration/imaging"
)

func (ip *ImageEditor) TrySplit(rotated bool) []*ImageEditor {
	dim := (*ip.Img).Bounds()
	w, h := dim.Dx(), dim.Dy()

	if (w > h) == (ip.targetW > ip.targetH) {
		return []*ImageEditor{ip}
	}

	if w <= ip.targetH &&
		h <= ip.targetW &&
		rotated {
		spread := image.Image(imaging.Rotate270(*ip.Img))
		ip.Img = nil
		return []*ImageEditor{ip.Copy(&spread, manga.SplitRotated)}
	}

	var processors []*ImageEditor

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
			ip.Copy(&pageOne, manga.SplitToLeft),
			ip.Copy(&pageTwo, manga.SplitToRight),
		)
	}

	spread := image.Image(imaging.Rotate270(*ip.Img))
	processors = append(processors, ip.Copy(&spread, manga.SplitRotated))

	ip.Img = nil
	return processors
}
