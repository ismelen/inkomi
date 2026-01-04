package manga

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

var ResampleBicubic = imaging.CatmullRom

const AutoCropThreshold = 0.015

type ComicPage struct {
	opts            Options
	width           int
	height          int
	originalMode    string
	img             image.Image
	fill            string
	rotated         bool
	orgPath         string
	targetPathStart string
	targetPathOrder string
	gamma           float64
	palette         []uint8
}

func NewComicPage(opts Options, mode string, path string, img image.Image, fill string) *ComicPage {
	profile := opts.ProfileData

	originalMode := "RGB"
	if _, ok := img.(*image.Gray); ok {
		originalMode = "L"
	} else if _, ok := img.(*image.Gray16); ok {
		originalMode = "L"
	}

	page := &ComicPage{
		opts:            opts,
		width:           profile.Width,
		height:          profile.Height,
		palette:         profile.Palette,
		gamma:           profile.Gamma,
		originalMode:    originalMode,
		img:             img,
		fill:            fill,
		rotated:         false,
		orgPath:         path,
		targetPathStart: strings.TrimSuffix(path, filepath.Ext(path)),
	}

	switch mode {
	case "N":
		page.targetPathOrder = "-kcc-x"
		break
	case "R":
		page.targetPathOrder = "-kcc-d"
		break
	case "S1":
		page.targetPathOrder = "-kcc-b"
		break
	case "S2":
		page.targetPathOrder = "-kcc-c"
		break
	}

	return page
}

func (t *ComicPage) CropPageNumber(power float64) {
	// Stub
}

func (t *ComicPage) CropMargin(power float64) {
	// Stub
}

func (t *ComicPage) OptimizeForDisplay(eraseRainbow bool, isColor bool) {
	// Stub
}


func (t *ComicPage) SaveToDir() error {
	flags := []string{}
	if t.rotated {
		flags = append(flags, "Rotated")
	}
	if t.fill != "white" {
		flags = append(flags, "BlackBackground")
	}

	_, err := t.saveWithCodec(
		t.img,
		t.targetPathStart+t.targetPathOrder)

	if err != nil {
		return err
	}

	if _, err := os.Stat(t.orgPath); err == nil {
		os.Remove(t.orgPath)
	}
	return nil
}

func (p *ComicPage) saveWithCodec(img image.Image, targetPath string) (string, error) {
	targetPath += ".jpg"
	err := imaging.Save(img, targetPath, imaging.JPEGQuality(85))
	if err != nil {
		return "", err
	}
	return targetPath, nil
}
