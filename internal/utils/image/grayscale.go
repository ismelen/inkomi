package ImageUtils

import (
	"ismelen/ermc/internal/manga/domain/MangaModels"

	"github.com/disintegration/imaging"
)

func ConvertToGrayscale(payload *MangaModels.PagePart) {
	(*payload.Image) = imaging.Grayscale((*payload.Image))
}
