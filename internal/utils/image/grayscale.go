package ImageUtils

import (
	manga "ismelen/ermc/internal/manga/logic/models"

	"github.com/disintegration/imaging"
)

func ConvertToGrayscale(payload *manga.PagePayload) {
	(*payload.Image) = imaging.Grayscale((*payload.Image))
}
