package imaging_test

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/imaging"
	"github.com/stretchr/testify/assert"
)

func filled(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
	return img
}

func TestCropBox_InkInsideMargins_ReturnsInkBounds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		bg, fg color.Color
	}{
		{"white page", color.White, color.Black},
		{"black page", color.Black, color.White},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			img := filled(100, 200, c.bg)
			ink := image.Rect(20, 30, 80, 170)
			draw.Draw(img, ink, &image.Uniform{C: c.fg}, image.Point{}, draw.Src)

			// Act
			box := imaging.CropBox(img, 0.02)

			// Assert
			assert.Equal(t, ink, box)
		})
	}
}

func TestCropBox_BlankPage_ReturnsBounds(t *testing.T) {
	t.Parallel()

	// Arrange
	img := filled(50, 50, color.White)

	// Act
	box := imaging.CropBox(img, 0.02)

	// Assert
	assert.Equal(t, img.Bounds(), box)
}

func TestAtkinsonDither_Gradient_QuantizesTo16Levels(t *testing.T) {
	t.Parallel()

	// Arrange
	img := image.NewRGBA(image.Rect(0, 0, 64, 8))
	for x := range 64 {
		for y := range 8 {
			v := uint8(x * 4)
			img.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}

	// Act
	imaging.AtkinsonDither(img)

	// Assert
	for i := 0; i < len(img.Pix); i += 4 {
		for c := range 3 {
			assert.Zero(t, img.Pix[i+c]%17, "pixel %d channel %d", i/4, c)
		}
	}
}

func TestApplyLevels_Grayscale_RemovesColorAndClipsPoints(t *testing.T) {
	t.Parallel()

	// Arrange
	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 5, G: 5, B: 5, A: 255})
	img.SetRGBA(1, 0, color.RGBA{R: 200, G: 40, B: 40, A: 255})
	img.SetRGBA(2, 0, color.RGBA{R: 250, G: 250, B: 250, A: 255})

	// Act
	imaging.ApplyLevels(img, imaging.Levels{Gamma: 1, BlackPoint: 0.06, WhitePoint: 0.95, Contrast: 1, Grayscale: true})

	// Assert
	assert.Equal(t, color.RGBA{A: 255}, img.RGBAAt(0, 0))
	mid := img.RGBAAt(1, 0)
	assert.Equal(t, mid.R, mid.G)
	assert.Equal(t, mid.G, mid.B)
	assert.Equal(t, color.RGBA{R: 255, G: 255, B: 255, A: 255}, img.RGBAAt(2, 0))
}

func TestToRGBA_OffsetImage_CopiesToOrigin(t *testing.T) {
	t.Parallel()

	// Arrange
	src := image.NewGray(image.Rect(10, 10, 14, 12))
	src.SetGray(10, 10, color.Gray{Y: 200})

	// Act
	rgba := imaging.ToRGBA(src)

	// Assert
	assert.Equal(t, image.Rect(0, 0, 4, 2), rgba.Rect)
	assert.Equal(t, color.RGBA{R: 200, G: 200, B: 200, A: 255}, rgba.RGBAAt(0, 0))
}
