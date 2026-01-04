package ImageUtils

import (
	manga "ismelen/ermc/internal/manga/logic/models"
	"math"

	"github.com/disintegration/imaging"
)

var filters = map[string]imaging.ResampleFilter{
	"bicubic": imaging.CatmullRom,
	"lanczos": imaging.Lanczos,
}

const AUTO_CROP_THRESHOLD = 0.015

func ResizeImage(payload *manga.PagePayload, stretchUpscaleMode int, targetW, targetH int) {
	bounds := (*payload.Image).Bounds()
	imgH, imgW := bounds.Dy(), bounds.Dx()

	ratioDevice := targetH / targetW
	ratioImg := imgH / imgW

	method := "lanczos"
	if imgW < targetW && imgH < targetH {
		method = "bicubic"
	}

	if stretchUpscaleMode == 1 {
		(*payload.Image) = imaging.Resize(
			*payload.Image,
			targetW,
			targetH,
			filters[method],
		)
	} else if method == "bicubic" &&
		!(stretchUpscaleMode == 2) {
		// pass
	} else {
		if math.Abs(float64(ratioImg-ratioDevice)) < AUTO_CROP_THRESHOLD {
			(*payload.Image) = imaging.Fill(
				*payload.Image,
				targetW,
				targetH,
				imaging.Center,
				filters[method],
			)
		} else {
			(*payload.Image) = imaging.Fit(
				*payload.Image,
				targetW,
				targetH,
				filters[method],
			)
		}
	}

	// currentW := (*payload.Image).Bounds().Dx()
	// currentH := (*payload.Image).Bounds().Dy()

	// if payload.TargetPathOrder == "-kcc-a" || payload.TargetPathOrder == "-kcc-d" {
	// 	// Removed KindleAZW3 check and NoRotate check logic
	// 	if currentW > targetW*2 || currentH > targetH {
	// 		(*payload.Image) = imaging.Fit((*payload.Image), targetW*2, targetH, imaging.Lanczos)
	// 	}
	// 	return
	// }

	// ratioDevice := float64(targetH) / float64(targetW)
	// ratioImage := float64(currentH) / float64(currentW)

	// var filter imaging.ResampleFilter
	// if currentW < targetW && currentH < targetH {
	// 	filter = ResampleBicubic
	// } else {
	// 	filter = imaging.Lanczos
	// }

	// if stretchUpscaleMode == 1 { // Stretching
	// 	(*payload.Image) = imaging.Resize((*payload.Image), targetW, targetH, filter)
	// } else if currentW < targetW && currentH < targetH && stretchUpscaleMode != 2 { // Upscaling
	// 	// pass
	// } else {
	// 	if math.Abs(ratioImage-ratioDevice) < AutoCropThreshold {
	// 		(*payload.Image) = imaging.Fill((*payload.Image), targetW, targetH, imaging.Center, filter)
	// 	} else {
	// 		(*payload.Image) = imaging.Fit((*payload.Image), targetW, targetH, filter)
	// 	}
	// }
}
