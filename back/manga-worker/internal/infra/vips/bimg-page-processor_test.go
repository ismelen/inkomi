package vips_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pagePNG draws a w x h page whose left half is left and right half is right.
func pagePNG(t *testing.T, w, h int, left, right color.Color) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, image.Rect(0, 0, w/2, h), &image.Uniform{C: left}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(w/2, 0, w, h), &image.Uniform{C: right}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func meanLuma(t *testing.T, data []byte) float64 {
	t.Helper()

	img, _, err := image.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	b := img.Bounds()
	sum := 0.0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sum += float64(color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y)
		}
	}
	return sum / float64(b.Dx()*b.Dy())
}

func newProcessor(t *testing.T) *vips.BimgPageProcessor {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return vips.NewBimgPageProcessor(ctx, 2)
}

func TestBimgPageProcessor_Process_SinglePage_FitsTargetAsJPEG(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := models.DefaultMangaConfig()
	cfg.TargetWidth, cfg.TargetHeight = 300, 400

	// Act
	pages, err := newProcessor(t).Process(context.Background(), pagePNG(t, 600, 900, color.Black, color.White), cfg, false)

	// Assert
	require.NoError(t, err)
	require.Len(t, pages, 1)
	assert.Equal(t, models.SplitNone, pages[0].Split)
	assert.Equal(t, models.ImageFormatJPEG, pages[0].Format)
	assert.LessOrEqual(t, pages[0].Width, 300)
	assert.LessOrEqual(t, pages[0].Height, 400)
	_, format, err := image.DecodeConfig(bytes.NewReader(pages[0].Data))
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
}

func TestBimgPageProcessor_Process_Spread_SplitsInReadingOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		rtl       bool
		wantSplit []models.SplitOperation
		firstDark bool
	}{
		{"ltr", false, []models.SplitOperation{models.SplitToLeft, models.SplitToRight}, true},
		{"rtl", true, []models.SplitOperation{models.SplitToRight, models.SplitToLeft}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cfg := models.DefaultMangaConfig()
			cfg.TargetWidth, cfg.TargetHeight = 300, 400
			spread := pagePNG(t, 1200, 800, color.Black, color.White)

			// Act
			pages, err := newProcessor(t).Process(context.Background(), spread, cfg, c.rtl)

			// Assert
			require.NoError(t, err)
			require.Len(t, pages, 2)
			assert.Equal(t, c.wantSplit, []models.SplitOperation{pages[0].Split, pages[1].Split})
			assert.Equal(t, c.firstDark, meanLuma(t, pages[0].Data) < 128)
		})
	}
}

func TestBimgPageProcessor_Process_SpreadWithRotate_ReturnsOneRotatedPage(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := models.DefaultMangaConfig()
	cfg.RotateDoublePage = true
	cfg.TargetWidth, cfg.TargetHeight = 300, 400

	// Act
	pages, err := newProcessor(t).Process(context.Background(), pagePNG(t, 1200, 800, color.Black, color.White), cfg, true)

	// Assert
	require.NoError(t, err)
	require.Len(t, pages, 1)
	assert.Equal(t, models.SplitRotated, pages[0].Split)
	assert.Greater(t, pages[0].Height, pages[0].Width)
}

func TestBimgPageProcessor_Process_Palette8bit_ReturnsPNG(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := models.DefaultMangaConfig()
	cfg.Palette8bit, cfg.Grayscale, cfg.Dithering = true, true, true
	cfg.CropTolerance = 0.02
	cfg.Sharpen.Enabled = true

	// Act
	pages, err := newProcessor(t).Process(context.Background(), pagePNG(t, 600, 900, color.Gray{Y: 90}, color.White), cfg, false)

	// Assert
	require.NoError(t, err)
	require.Len(t, pages, 1)
	assert.Equal(t, models.ImageFormatPNG, pages[0].Format)
	_, format, err := image.DecodeConfig(bytes.NewReader(pages[0].Data))
	require.NoError(t, err)
	assert.Equal(t, "png", format)
}

func TestBimgPageProcessor_Process_NotAnImage_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	_, err := newProcessor(t).Process(context.Background(), []byte("not an image"), models.DefaultMangaConfig(), false)

	// Assert
	assert.Error(t, err)
}

func TestBimgPageProcessor_Process_CanceledContext_ReturnsContextError(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	_, err := newProcessor(t).Process(ctx, pagePNG(t, 60, 90, color.Black, color.White), models.DefaultMangaConfig(), false)

	// Assert
	assert.ErrorIs(t, err, context.Canceled)
}
