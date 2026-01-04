package PageConverter

import (
	"fmt"
	manga "ismelen/ermc/internal/manga/logic/models"
	FileUtils "ismelen/ermc/internal/utils/file"
	ImageUtils "ismelen/ermc/internal/utils/image"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type PageConverter struct {
	page *manga.PageData
	opts *manga.Options
	cPayload *manga.PagePayload
	chapterName string
}

func New(page *manga.PageData, opts *manga.Options, chapterName string) (*PageConverter, error) {
	img, err := imaging.Open(page.Src)
	if err != nil {
		return nil, err
	}

	width, height := opts.ProfileData.Width, opts.ProfileData.Height

	page.Img = &img
	page.Fill = ImageUtils.FillCheck(
		width,
		height,
		&img,
	)

	ImageUtils.SplitCheck(page, width, height, opts)

	return &PageConverter{
		page: page,
		opts: opts,
		chapterName: chapterName,
	}, nil
}

func (t *PageConverter) Convert(pageNum int) error {
	pathParts := strings.Split(t.page.Src, string(filepath.Separator))
	filePath := filepath.Join(pathParts[len(pathParts)-2:]...)
	fmt.Printf("Processing: %s\n", filePath)
	
	for _, payload := range t.page.Payloads {
		t.cPayload = payload
		if err := t.ConvertPayload(pageNum); err != nil {
			return err
		}
	}
	t.page.Img = nil
	runtime.GC()
	return nil
}

func (t *PageConverter) ConvertPayload(pageNum int) error{
	switch t.opts.CroppingMode {
	case 2:
		ImageUtils.CropPageNumber(
			t.cPayload,
			float64(t.opts.CroppingPower),
			t.page.Fill,
			0.0,
			t.opts.PreserveMargin,
		)
	case 1:
		ImageUtils.CropMargins(
			t.cPayload,
			float64(t.opts.CroppingPower),
			t.page.Fill,
			0.0,
			t.opts.PreserveMargin,
		)
	}
	
	ImageUtils.ResizeImage(
		t.cPayload,
		t.opts.StretchUpscaleMode,
		t.opts.ProfileData.Width,
		t.opts.ProfileData.Height,
	)

	isColor := t.opts.ColorMode && t.isColor()
	if !isColor {
		ImageUtils.ConvertToGrayscale(t.cPayload)
	}

	if t.opts.RainbowEraser {
		ImageUtils.OptimizeForDisplay(t.cPayload, isColor)
	}
	
	return t.SaveToDir(pageNum)
	
}


func (t *PageConverter) isColor() bool {
	if t.cPayload.OriginalMode == "L" {
		return false
	}

	if ImageUtils.CalculateColor(t.opts.ColorMode, t.cPayload.Image) {
		return true
	}

	return true
}

func (t *PageConverter) SaveToDir(pageNum int) error {
	title := fmt.Sprintf("ermc-%d%s", pageNum, t.cPayload.TargetPathOrder)
	path, err := FileUtils.SaveWithCodec(
		t.cPayload.Image,
		filepath.Join(
			t.opts.Output,
			"chapters",
			t.chapterName,
			title,
		),
	)

	t.cPayload.Path = path
	t.cPayload.Title = title

	dim := (*t.cPayload.Image).Bounds()
	t.cPayload.H = dim.Dy()
	t.cPayload.W = dim.Dx()
	
	t.cPayload.Image = nil

	return err
}
