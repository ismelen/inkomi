package image

import "github.com/disintegration/imaging"

func (ip *ImageProcessor) Resize() {
	tH, tW := ip.targetH, ip.targetW
	wRatio, hRatio := ip.getImgRatios()

	if wRatio < hRatio {
		tH = 0
	} else {
		tW = 0
	}

	filter := imaging.Lanczos
	if ip.isImgSmaller() {
		filter = imaging.CatmullRom
	}

	(*ip.Img) = imaging.Resize(
		*ip.Img,
		tW,
		tH,
		filter,
	)
}

func (ip *ImageProcessor) getImgRatios() (float64, float64) {
	bounds := (*ip.Img).Bounds()
	imgH, imgW := bounds.Dy(), bounds.Dx()
	tH, tW := ip.targetH, ip.targetW

	wRatio := float64(tW) / float64(imgW)
	hRatio := float64(tH) / float64(imgH)

	return wRatio, hRatio
}

func (ip *ImageProcessor) isImgSmaller() bool {
	bounds := (*ip.Img).Bounds()
	imgH, imgW := bounds.Dy(), bounds.Dx()

	return imgW < ip.targetW && imgH < ip.targetH
}
