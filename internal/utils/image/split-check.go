package ImageUtils

import (
	"image"
	manga "ismelen/ermc/internal/manga/logic/models"

	"github.com/disintegration/imaging"
)

func SplitCheck(page *manga.PageData, dstWidth, dstHeight int, opts *manga.Options) {
	width := (*page.Img).Bounds().Dx()
	height := (*page.Img).Bounds().Dy()

	if (width > height) != (dstWidth > dstHeight) {
		if width <= dstHeight &&
			height <= dstWidth &&
			opts.SpreadSplitter == 2 {

			spread := (*page.Img)
			spread = imaging.Rotate90(spread)
			page.Payloads = append(page.Payloads, manga.NewPagePayload("R", &spread))
		} else {
			if opts.SpreadSplitter != 2 {
				var leftBox, rightBox image.Rectangle
				if width < height {
					leftBox = image.Rect(0, 0, width, height/2)
					rightBox = image.Rect(0, width/2, width, height)
				} else {
					leftBox = image.Rect(0, 0, width/2, height)
					rightBox = image.Rect(width/2, 0, width, height)
				}

				var pageOne, pageTwo image.Image
				if opts.Manga {
					pageOne = imaging.Crop((*page.Img), rightBox)
					pageTwo = imaging.Crop((*page.Img), leftBox)
				} else {
					pageOne = imaging.Crop((*page.Img), leftBox)
					pageTwo = imaging.Crop((*page.Img), rightBox)
				}

				page.Payloads = append(page.Payloads, manga.NewPagePayload("S1", &pageOne))
				page.Payloads = append(page.Payloads, manga.NewPagePayload("S2", &pageTwo))
			}

			if opts.SpreadSplitter == 1 {
				spread := (*page.Img)
				spread = imaging.Rotate90(spread)
				page.Payloads = append(page.Payloads, manga.NewPagePayload("R", &spread))
			}
		}
	} else {
		page.Payloads = append(page.Payloads, manga.NewPagePayload("N", page.Img))
	}
}
