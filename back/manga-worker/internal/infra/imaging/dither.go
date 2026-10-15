package imaging

import (
	"image"
	"math"
)

// ditherStep reduces each channel to 16 levels (255/15), the depth of 16-grey e-ink screens.
const ditherStep float32 = 17.0

// atkinson are the (dx, dy) neighbours that get 1/8 of the quantization error each: two to
// the right, three on the next row and one two rows below.
var atkinson = [...]struct{ dx, dy int }{{1, 0}, {2, 0}, {-1, 1}, {0, 1}, {1, 1}, {0, 2}}

// AtkinsonDither applies Atkinson error diffusion in place, reducing each channel to 16 levels.
func AtkinsonDither(img *image.RGBA) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	errs := [3][]float32{
		make([]float32, width*height),
		make([]float32, width*height),
		make([]float32, width*height),
	}

	for y := range height {
		for x := range width {
			idx := y*width + x
			offset := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)

			for c := range 3 {
				old := float32(img.Pix[offset+c]) + errs[c][idx]
				quantized := closestLevel(old)
				img.Pix[offset+c] = quantized

				spread := (old - float32(quantized)) / 8
				for _, n := range atkinson {
					tx, ty := x+n.dx, y+n.dy
					if tx >= 0 && tx < width && ty < height {
						errs[c][ty*width+tx] += spread
					}
				}
			}
		}
	}
}

func closestLevel(v float32) uint8 {
	v = min(max(v, 0), 255)
	return uint8(math.Round(float64(v/ditherStep)) * float64(ditherStep))
}
