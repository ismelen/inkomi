package imaging

import (
	"image"
	"image/color"
	"math"
)

type Levels struct {
	Gamma      float64 // 1.0 is neutral
	BlackPoint float64 // [0, 1]: lumas below become pure black
	WhitePoint float64 // [0, 1]: lumas above become pure white
	Brightness float64 // [-1, 1]
	Contrast   float64 // [0, 2], 1.0 is neutral
	Grayscale  bool
}

// saturationBoost compensates the washed out colours of colour e-ink screens.
const saturationBoost = 1.5

// ApplyLevels stretches the luminance between its 1% and 99% percentiles, clips it at the
// black and white points, applies contrast, brightness and gamma, and either drops or boosts
// the saturation. It works in place.
func ApplyLevels(img *image.RGBA, opts Levels) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return
	}

	gamma := opts.Gamma
	if gamma == 0 {
		gamma = 1
	}
	contrast := opts.Contrast
	if contrast == 0 {
		contrast = 1
	}

	histogram := make([]int, 256)
	pixels := make([]hsl, width*height)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			p := rgbToHSL(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255)
			pixels[idx] = p
			idx++
			histogram[lumaIndex(p.l)]++
		}
	}

	p1, p99 := percentiles(histogram, width*height)

	lut := make([]float64, 256)
	for i := range lut {
		v := float64(i) / 255
		switch {
		case v <= opts.BlackPoint, i <= p1:
			lut[i] = 0
		case opts.WhitePoint > 0 && v >= opts.WhitePoint, i >= p99:
			lut[i] = 1
		default:
			norm := float64(i-p1) / float64(p99-p1)
			norm = (norm-0.5)*contrast + 0.5 + opts.Brightness
			lut[i] = math.Pow(min(max(norm, 0), 1), gamma)
		}
	}

	idx = 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := pixels[idx]
			idx++

			s := 0.0
			if !opts.Grayscale {
				s = min(p.s*saturationBoost, 1)
			}
			r, g, b := hslToRGB(p.h, s, lut[lumaIndex(p.l)])
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(math.Round(r * 255)),
				G: uint8(math.Round(g * 255)),
				B: uint8(math.Round(b * 255)),
				A: 255,
			})
		}
	}
}

// percentiles returns the 1% and 99% luma percentiles, or the full range when the image is
// so flat that they collapse.
func percentiles(histogram []int, total int) (p1, p99 int) {
	p1Target := int(float64(total) * 0.01)
	p99Target := int(float64(total) * 0.99)

	count := 0
	p1Found := false
	p99 = 255
	for i, n := range histogram {
		count += n
		if !p1Found && count >= p1Target {
			p1, p1Found = i, true
		}
		if count >= p99Target {
			p99 = i
			break
		}
	}
	if p99 <= p1 {
		return 0, 255
	}
	return p1, p99
}

func lumaIndex(l float64) int {
	return min(max(int(math.Round(l*255)), 0), 255)
}
