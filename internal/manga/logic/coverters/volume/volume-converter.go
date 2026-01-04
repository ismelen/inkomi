package VolumeConverter

import (
	"fmt"
	EpubBuilder "ismelen/ermc/internal/manga/logic/builders/epub"
	manga "ismelen/ermc/internal/manga/logic/models"
	FileUtils "ismelen/ermc/internal/utils/file"
	ImageUtils "ismelen/ermc/internal/utils/image"
	"os"
	"path/filepath"
	"runtime"
)

type MangaPageConverter struct {
	opts manga.Options
}

func NewVolumeConverter(opts manga.Options) *MangaPageConverter {
	return &MangaPageConverter{opts: opts}
}

func Convert(opts *manga.Options, dstFileName string, chapters ...*manga.ChapterData) (string, error) {
	var pageNum int

	for _, chapter := range chapters {
		for _, page := range chapter.Pages {
			pageNum++
			fmt.Printf("Processing: %s\n", page.Src)
			for _, payload := range page.Payloads {

				isColor := Color(opts.ColorMode, payload) && opts.ColorMode

				if opts.CroppingMode == 2 {
					ImageUtils.CropPageNumber(
						payload,
						float64(opts.CroppingPower),
						page.Fill,
						0.0,
						opts.PreserveMargin,
					)
				}

				if opts.CroppingMode == 1 {
					ImageUtils.CropMargins(
						payload,
						float64(opts.CroppingPower),
						page.Fill,
						0.0,
						opts.PreserveMargin,
					)
				}

				ImageUtils.ResizeImage(
					payload,
					opts.StretchUpscaleMode,
					opts.ProfileData.Width,
					opts.ProfileData.Height,
				)

				if opts.RainbowEraser {
					ImageUtils.OptimizeForDisplay(payload, isColor)
				}

				if !isColor {
					ImageUtils.ConvertToGrayscale(payload)
				}

				// Save to dir
				title := fmt.Sprintf("ermc-%d%s", pageNum, payload.TargetPathOrder)
				path, err := FileUtils.SaveWithCodec(
					payload.Image,
					filepath.Join(
						opts.Output,
						"chapters",
						chapter.NormalizedName,
						title,
					),
				)
				payload.Image = nil
				payload.Path = path
				payload.Title = title
				if err != nil {
					return "", err
				}

				if _, err := os.Stat(page.Src); err == nil {
					os.Remove(page.Src)
				}

				runtime.GC()
			}
		}
	}

	// Generate output
	return generateOutput(opts, dstFileName, chapters...)
}

func generateOutput(opts *manga.Options, dstFileName string, chapters ...*manga.ChapterData) (path string, err error) {
	switch opts.Format {
	case "Auto", "CBZ", "PDF", "EPUB":
		builder := EpubBuilder.New(opts, dstFileName, chapters...)
		path, err = builder.Build()
	}

	if path == "" {
		return "", fmt.Errorf("Cannot generate output for %s", dstFileName)
	}

	return
}

func Color(colorMode bool, payload *manga.PagePayload) bool {
	if payload.OriginalMode == "L" {
		return false
	}

	if ImageUtils.CalculateColor(colorMode, payload.Image) {
		return true
	}

	return true
}

// func (t *MangaPageConverter) removeNonImages(source string) {
// 	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
// 		if !info.IsDir() && !utils.IsImage(path) {
// 			os.Remove(path)
// 		}
// 		return nil
// 	})
// }
