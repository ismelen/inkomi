package imaging_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/imaging"
	"github.com/stretchr/testify/assert"
)

// stripes alternates two colours of the same luma every column, like the colour fringes of a
// screentone on a colour screen.
func stripes(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	a := color.YCbCr{Y: 128, Cb: 80, Cr: 180}
	b := color.YCbCr{Y: 128, Cb: 176, Cr: 76}
	for y := range height {
		for x := range width {
			c := a
			if x%2 == 1 {
				c = b
			}
			r, g, bl := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: bl, A: 255})
		}
	}
	return img
}

func chroma(c color.RGBA) float64 {
	_, cb, cr := color.RGBToYCbCr(c.R, c.G, c.B)
	return math.Hypot(float64(cb)-128, float64(cr)-128)
}

func TestRemoveRainbow_FineStripes_ChromaFadesLumaStays(t *testing.T) {
	t.Parallel()

	// Arrange
	img := stripes(20, 10)
	before := chroma(img.RGBAAt(10, 5))

	// Act
	imaging.RemoveRainbow(img, imaging.RainbowRadius)

	// Assert
	after := img.RGBAAt(10, 5)
	assert.Less(t, chroma(after), before/3, "the alternating colour averages out")
	l, _, _ := color.RGBToYCbCr(after.R, after.G, after.B)
	assert.InDelta(t, 128, int(l), 3, "the brightness of the drawing is kept")
}

func TestRemoveRainbow_FlatColour_Unchanged(t *testing.T) {
	t.Parallel()

	// Arrange: a solid colour has no fringes to remove
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}

	// Act
	imaging.RemoveRainbow(img, imaging.RainbowRadius)

	// Assert
	got := img.RGBAAt(4, 4)
	assert.InDelta(t, 200, int(got.R), 3)
	assert.InDelta(t, 40, int(got.G), 3)
	assert.InDelta(t, 40, int(got.B), 3)
}
