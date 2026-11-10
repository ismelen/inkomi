package imaging

import (
	"image"
	"image/color"
)

// RainbowRadius is the chroma blur radius, the one the legacy api used.
const RainbowRadius = 2

// RemoveRainbow softens the colour fringes colour e-ink screens show on fine screentones and
// sharp edges: every pixel keeps its own luma, so the drawing stays sharp, and takes its chroma
// from a box blur of radius pixels around it. It works in place.
func RemoveRainbow(img *image.RGBA, radius int) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 || radius <= 0 {
		return
	}

	lumas := make([]uint8, width*height)
	cb := make([]float32, width*height)
	cr := make([]float32, width*height)
	for y := range height {
		for x := range width {
			c := img.RGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
			l, b, r := color.RGBToYCbCr(c.R, c.G, c.B)
			i := y*width + x
			lumas[i], cb[i], cr[i] = l, float32(b), float32(r)
		}
	}

	boxBlur(cb, width, height, radius)
	boxBlur(cr, width, height, radius)

	for y := range height {
		for x := range width {
			i := y*width + x
			r, g, b := color.YCbCrToRGB(lumas[i], clampByte(cb[i]), clampByte(cr[i]))
			img.SetRGBA(bounds.Min.X+x, bounds.Min.Y+y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

// boxBlur averages each value with its neighbours up to radius away, horizontally and then
// vertically, in O(width × height) whatever the radius. Edges average the pixels inside.
func boxBlur(values []float32, width, height, radius int) {
	tmp := make([]float32, len(values))
	blurLines(values, tmp, height, width, 1, width, radius)
	blurLines(tmp, values, width, height, width, 1, radius)
}

// blurLines blurs lines of length size from src into dst; step moves along a line and next to
// the following line.
func blurLines(src, dst []float32, lines, size, step, next, radius int) {
	for line := range lines {
		base := line * next
		var sum float32
		count := 0
		for i := 0; i <= radius && i < size; i++ {
			sum += src[base+i*step]
			count++
		}
		for i := range size {
			dst[base+i*step] = sum / float32(count)
			if out := i - radius; out >= 0 {
				sum -= src[base+out*step]
				count--
			}
			if in := i + radius + 1; in < size {
				sum += src[base+in*step]
				count++
			}
		}
	}
}

func clampByte(v float32) uint8 {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	default:
		return uint8(v + 0.5)
	}
}
