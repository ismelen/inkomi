package ImageUtils

import "ismelen/ermc/internal/manga/domain/MangaModels"

func OptimizeForDisplay(payload *MangaModels.PagePart) {
	bounds := (*payload.Image).Bounds()
	if bounds.Dx() > 1 && bounds.Dy() > 1 {
		EraseRainbowArtifacts(payload.Image)
	}
}
