package imagev2

import (
	"image"
	"image/color"
	"math"
)

// HSL represents a color in HSL space
type HSL struct {
	H, S, L float64
}

// applyLevelsAndColor normalizes contrast based on luminance percentiles
// and boosts saturation.
func applyLevelsAndColor(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create histogram for Luminance
	histogram := make([]int, 256)
	hslData := make([]HSL, width*height)

	// Single pass: convert to HSL and build histogram
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0

			h, s, l := rgbToHSL(r, g, b)
			hslData[idx] = HSL{h, s, l}
			idx++

			lumaIdx := int(math.Round(l * 255.0))
			if lumaIdx < 0 {
				lumaIdx = 0
			}
			if lumaIdx > 255 {
				lumaIdx = 255
			}
			histogram[lumaIdx]++
		}
	}

	// Calculate 1% and 99% percentiles
	totalPixels := width * height
	p1Target := int(float64(totalPixels) * 0.01)
	p99Target := int(float64(totalPixels) * 0.99)

	var p1, p99 int
	count := 0
	for i := 0; i < 256; i++ {
		count += histogram[i]
		if p1 == 0 && count >= p1Target {
			p1 = i
		}
		if count >= p99Target {
			p99 = i
			break
		}
	}
	if p99 == 0 {
		p99 = 255
	}

	// Generate LUT for Luminance with a slight gamma curve (e.g., gamma = 0.9 to lighten midtones)
	lut := make([]float64, 256)
	gamma := 0.9

	for i := 0; i < 256; i++ {
		if i <= p1 {
			lut[i] = 0.0
		} else if i >= p99 {
			lut[i] = 1.0
		} else {
			// Normalize to 0.0 - 1.0 range based on percentiles
			norm := float64(i-p1) / float64(p99-p1)
			// Apply gamma
			lut[i] = math.Pow(norm, gamma)
		}
	}

	// Apply corrections: mapping luminance and boosting saturation
	idx = 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			hsl := hslData[idx]
			idx++

			// Apply L stretching
			lumaIdx := int(math.Round(hsl.L * 255.0))
			if lumaIdx < 0 {
				lumaIdx = 0
			}
			if lumaIdx > 255 {
				lumaIdx = 255
			}
			newL := lut[lumaIdx]

			// Boost S
			newS := hsl.S * 1.5
			if newS > 1.0 {
				newS = 1.0
			}

			// Back to RGB
			r, g, b := hslToRGB(hsl.H, newS, newL)
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(math.Round(r * 255.0)),
				G: uint8(math.Round(g * 255.0)),
				B: uint8(math.Round(b * 255.0)),
				A: 255,
			})
		}
	}
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	l = (max + min) / 2.0

	if max == min {
		h = 0
		s = 0 // achromatic
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6.0
			}
		case g:
			h = (b-r)/d + 2.0
		case b:
			h = (r-g)/d + 4.0
		}
		h /= 6.0
	}
	return
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q

	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)
	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0.0 {
		t += 1.0
	}
	if t > 1.0 {
		t -= 1.0
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6.0*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}
