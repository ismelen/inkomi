package PageConverter

import (
	"fmt"
	"image"
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
	chapterName string
}

func New(page *manga.PageData, opts *manga.Options, chapterName string) (*PageConverter, error) {
	img, err := imaging.Open(page.Src)
	if err != nil {
		return nil, err
	}
	
	page.BgColor = ImageUtils.FillCheck(&img)
	
	tW, tH := opts.ProfileData.Width, opts.ProfileData.Height
	payloads, count := ImageUtils.SplitCheck(&img, tW, tH, opts)
	page.Count = count
	page.Payloads = payloads

	pathParts := strings.Split(page.Src, string(filepath.Separator))
	filePath := filepath.Join(pathParts[len(pathParts)-2:]...)
	fmt.Printf("Processing: %s\n", filePath)

	page.Src = ""
	runtime.GC()
	
	return &PageConverter{
		page: page,
		opts: opts,
		chapterName: chapterName,
	}, nil
}

func (t *PageConverter) Convert(pageNum int) error {
	var i int8
	for ; i<t.page.Count; i++ {
		if err := t.ConvertPayload(pageNum, t.page.Payloads[i]); err != nil {
			return err
		}
	}

	return nil
}

func (t *PageConverter) ConvertPayload(pageNum int, payload *manga.PagePayload) error{
	switch t.opts.CroppingMode {
	case 2:
		ImageUtils.CropPageNumber(
			payload,
			t.page.BgColor,
			t.opts.PreserveMargin,
		)
	case 1:
		ImageUtils.CropMargins(
			payload,
			t.page.BgColor,
			t.opts.PreserveMargin,
		)
	}

	ImageUtils.ResizeImage(
		payload,
		t.opts.StretchUpscaleMode,
		t.opts.ProfileData.Width,
		t.opts.ProfileData.Height,
	)

	isColor := t.opts.ColorMode && t.isColor(payload)
	fmt.Println(isColor)
	if !isColor {
		ImageUtils.ConvertToGrayscale(payload)
	}

	if t.opts.RainbowEraser {
		ImageUtils.OptimizeForDisplay(payload)
	}
	
	return t.SaveToDir(pageNum, payload)
	
}


func (t *PageConverter) isColor(payload *manga.PagePayload) bool {

	switch (*payload.Image).(type) {
	case *image.Gray:
		return false
	}
	
	detector := ImageUtils.ColorDetector{
		ForceColor: t.opts.ColorMode,
	}

	return detector.CalculateColor(*payload.Image)
}

func (t *PageConverter) SaveToDir(pageNum int, payload *manga.PagePayload) error {
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
