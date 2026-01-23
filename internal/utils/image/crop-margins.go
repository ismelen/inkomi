package ImageUtils

import (
	"image"
	MangaModels "ismelen/ermc/internal/manga/domain/models"

	"github.com/disintegration/imaging"
)

func CropPageNumber(payload *MangaModels.PagePart, hasWhiteBg bool, preserveMargin float64) {
	cropFunc(payload, hasWhiteBg, preserveMargin, CalculateBboxAgresive)
}

func CropMargins(payload *MangaModels.PagePart, hasWhiteBg bool, preserveMargin float64) {
	cropFunc(payload, hasWhiteBg, preserveMargin, CalculateBbox)
}

func cropFunc(payload *MangaModels.PagePart, hasWhiteBg bool, preserveMargin float64, cropper func (img image.Image, hasWhiteBg bool) BBox) {
	img := image.Image(imaging.AdjustContrast(*payload.Image, 100))
	img = imaging.Grayscale(img)
	
	box := cropper(img, hasWhiteBg)

	rect := image.Rect(
		box.Left - int(preserveMargin),
		box.Top - int(preserveMargin),
		box.Right + int(preserveMargin),
		box.Bottom + int(preserveMargin),
	)

	(*payload.Image) = imaging.Crop(*payload.Image, rect)
}
