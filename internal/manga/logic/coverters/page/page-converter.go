package PageConverter

import (
	"fmt"
	"image"
	"ismelen/ermc/internal/manga/domain/MangaModels"
	FileUtils "ismelen/ermc/internal/utils/file"
	ImageUtils "ismelen/ermc/internal/utils/image"
	"path/filepath"
	"runtime"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type PageConverter struct {
	page *MangaModels.Page
	opts *MangaModels.ConverterOptions
	chapterName string
}

func New(page *MangaModels.Page, opts *MangaModels.ConverterOptions) (*PageConverter, error) {
	img, err := imaging.Open(page.Path)
	if err != nil {
		return nil, err
	}
	
	page.HasWhiteBg = ImageUtils.HasWhiteBg(&img)
	payloads, count := ImageUtils.SplitCheck(
		&img, 
		opts.ProfileData.Width, 
		opts.ProfileData.Height, 
		opts,
	)
	page.Count = count
	page.Parts = payloads

	dir, _ := filepath.Split(page.Path)
	page.Path = ""
	runtime.GC()


	return &PageConverter{
		page: page,
		opts: opts,
		chapterName: filepath.Base(dir),
	}, nil
}

func (t *PageConverter) Convert(pageNum int) error {
	for i := range t.page.Count {
		if err := t.ConvertPayload(pageNum, t.page.Parts[i]); err != nil {
			return err
		}
	}

	return nil
}

func (t *PageConverter) ConvertPayload(pageNum int, part *MangaModels.PagePart) error{
	switch t.opts.CroppingMode {
	case 2:
		ImageUtils.CropPageNumber(
			part,
			t.page.HasWhiteBg,
			t.opts.PreserveMargin,
		)
	case 1:
		ImageUtils.CropMargins(
			part,
			t.page.HasWhiteBg,
			t.opts.PreserveMargin,
		)
	}

	ImageUtils.ResizeImage(
		part,
		t.opts.StretchUpscaleMode,
		t.opts.ProfileData.Width,
		t.opts.ProfileData.Height,
	)

	isColor := t.opts.ColorMode && t.isColor(part)
	if !isColor {
		ImageUtils.ConvertToGrayscale(part)
	}

	if t.opts.RainbowEraser {
		ImageUtils.OptimizeForDisplay(part)
	}
	
	return t.SaveToDir(pageNum, part)
	
}


func (t *PageConverter) isColor(payload *MangaModels.PagePart) bool {
	switch (*payload.Image).(type) {
	case *image.Gray:
		return false
	}
	
	detector := ImageUtils.ColorDetector{
		ForceColor: t.opts.ColorMode,
	}

	return detector.CalculateColor(*payload.Image)
}

func (t *PageConverter) SaveToDir(pageNum int, payload *MangaModels.PagePart) error {
	title := fmt.Sprintf("ermc-%d%s", pageNum, payload.TargetPathOrder)
	path, err := FileUtils.SaveWithCodec(
		payload.Image,
		filepath.Join(
			t.opts.Output,
			"chapters",
			t.chapterName,
			title,
		),
	)

	payload.Path = path
	payload.Title = title

	dim := (*payload.Image).Bounds()
	payload.H = dim.Dy()
	payload.W = dim.Dx()
	
	payload.Image = nil
	runtime.GC()

	return err
}
