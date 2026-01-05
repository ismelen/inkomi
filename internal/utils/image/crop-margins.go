package ImageUtils

import (
	"image"
	manga "ismelen/ermc/internal/manga/logic/models"

	"github.com/disintegration/imaging"
)

func CropPageNumber(payload *manga.PagePayload, bgColor string, preserveMargin float64) {
	cropFunc(payload, bgColor, preserveMargin, CalculateBboxAgresive)
}

func CropMargins(payload *manga.PagePayload, bgColor string, preserveMargin float64) {
	cropFunc(payload, bgColor, preserveMargin, CalculateBbox)
}

func cropFunc(payload *manga.PagePayload, bgColor string, preserveMargin float64, cropper func (img image.Image, bgColor string) BBox) {
	img := image.Image(imaging.AdjustContrast(*payload.Image, 100))
	img = imaging.Grayscale(img)
	
	box := cropper(img, bgColor)

	rect := image.Rect(
		box.Left - int(preserveMargin),
		box.Top - int(preserveMargin),
		box.Right + int(preserveMargin),
		box.Bottom + int(preserveMargin),
	)

	(*payload.Image) = imaging.Crop(*payload.Image, rect)
}
