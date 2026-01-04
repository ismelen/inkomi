package ImageUtils

import manga "ismelen/ermc/internal/manga/logic/models"

func OptimizeForDisplay(payload *manga.PagePayload, isColor bool) {
	bounds := (*payload.Image).Bounds()
	if bounds.Dx() > 1 && bounds.Dy() > 1 {
		EraseRainbowArtifacts(payload.Image, isColor)
	}
}
