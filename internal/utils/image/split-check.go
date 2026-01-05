package ImageUtils

import (
	"image"
	manga "ismelen/ermc/internal/manga/logic/models"

	"github.com/disintegration/imaging"
)

func SplitCheck(img *image.Image, tw, th int, opts *manga.Options) (payloads [3]*manga.PagePayload, count int8) {
	w := (*img).Bounds().Dx()
	h := (*img).Bounds().Dy()

	if (w > h) == (tw > th) {
		// page.Payloads = append(page.Payloads, manga.NewPagePayload('N', img))
		payloads[count] = manga.NewPagePayload('N', img)
		count++
		return
	}

	if w <= th && 
		h <= tw && 
		opts.SpreadSplitter == 2 {
		spread := image.Image(imaging.Rotate270(*img))
		// page.Payloads = append(page.Payloads, manga.NewPagePayload(
		// 	'R', 
		// 	&spread,
		// ))
		payloads[count] = manga.NewPagePayload('R', &spread)
		count++
		return
	}

	if opts.SpreadSplitter != 2 {
		var leftBox, rightBox image.Rectangle
		if w < h {
			leftBox = image.Rect(0, 0, w, h/2)
			rightBox = image.Rect(0, w/2, w, h)
		} else {
			leftBox = image.Rect(0, 0, w/2, h)
			rightBox = image.Rect(w/2, 0, w, h)
		}

		var pageOne, pageTwo image.Image
		if opts.Manga {
			pageOne = imaging.Crop(*img, rightBox)
			pageTwo = imaging.Crop(*img, leftBox)
		} else {
			pageOne = imaging.Crop(*img, leftBox)
			pageTwo = imaging.Crop(*img, rightBox)
		}

		payloads[count] = manga.NewPagePayload('1', &pageOne)
		payloads[count+1] = manga.NewPagePayload('2', &pageTwo)
		count += 2
	}

	if opts.SpreadSplitter == 1 {
		spread := image.Image(imaging.Rotate270(*img))
		payloads[count] = manga.NewPagePayload('R', &spread)
		count++
	}

	return
}
