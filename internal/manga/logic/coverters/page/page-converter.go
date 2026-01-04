package PageConverter

import (
	manga "ismelen/ermc/internal/manga/logic/models"
	ImageUtils "ismelen/ermc/internal/utils/image"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

func Convert(page *manga.PageData, opts *manga.Options) error {
	img, err := imaging.Open(page.Src)
	if err != nil {
		return err
	}

	width, height := opts.ProfileData.Width, opts.ProfileData.Height

	page.Img = &img
	page.Fill = ImageUtils.FillCheck(
		width,
		height,
		&img,
	)

	ImageUtils.SplitCheck(page, width, height, opts)

	return nil
}
