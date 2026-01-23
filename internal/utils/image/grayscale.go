package ImageUtils

import (
	MangaModels "ismelen/ermc/internal/manga/domain/models"

	"github.com/disintegration/imaging"
)

func ConvertToGrayscale(payload *MangaModels.PagePart) {
	(*payload.Image) = imaging.Grayscale((*payload.Image))
}
