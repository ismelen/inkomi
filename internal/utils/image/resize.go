package ImageUtils

import (
	manga "ismelen/ermc/internal/manga/logic/models"

	"github.com/disintegration/imaging"
)

var filters = map[string]imaging.ResampleFilter{
	"bicubic": imaging.CatmullRom,
	"lanczos": imaging.Lanczos,
}

const AUTO_CROP_THRESHOLD = 0.015

func ResizeImage(payload *manga.PagePayload, stretchUpscaleMode int, tW, tH int) {
	
	bounds := (*payload.Image).Bounds()
	imgH, imgW := bounds.Dy(), bounds.Dx()

	method := "lanczos"
	if imgW < tW && imgH < tH {
		method = "bicubic"
	}

	ratioW := float64(tW) / float64(imgW)
	ratioH := float64(tH) / float64(imgH)

	if ratioW < ratioH {
		tH = 0
	} else{
		tW = 0
	}

	(*payload.Image) = imaging.Resize(
		*payload.Image, 
		tW, 
		tH, 
		filters[method],
	)
}
