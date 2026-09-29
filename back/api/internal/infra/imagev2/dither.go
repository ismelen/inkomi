package imagev2

import (
	"image"
	"math"
)

// applyAtkinsonDither applies Atkinson error diffusion dithering
// to reduce each channel to 16 levels (4096 colors total, matching Kaleido 3 E-Ink).
func applyAtkinsonDither(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create flat buffers for accumulated errors to avoid re-allocating inside the loop.
	errR := make([]float32, width*height)
	errG := make([]float32, width*height)
	errB := make([]float32, width*height)

	// 16 levels means step size is 255 / 15 = 17
	const step float32 = 17.0

	closestLevel := func(val float32) uint8 {
		if val < 0 {
			val = 0
		}
		if val > 255 {
			val = 255
		}
		level := math.Round(float64(val) / float64(step))
		return uint8(level * float64(step))
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			// Get current pixel with accumulated error
			offset := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)

			oldR := float32(img.Pix[offset]) + errR[idx]
			oldG := float32(img.Pix[offset+1]) + errG[idx]
			oldB := float32(img.Pix[offset+2]) + errB[idx]

			newR := closestLevel(oldR)
			newG := closestLevel(oldG)
			newB := closestLevel(oldB)

			img.Pix[offset] = newR
			img.Pix[offset+1] = newG
			img.Pix[offset+2] = newB

			quantErrorR := oldR - float32(newR)
			quantErrorG := oldG - float32(newG)
			quantErrorB := oldB - float32(newB)

			// Atkinson weights: 1/8
			distributeError := func(dx, dy int, er, eg, eb float32) {
				if x+dx >= 0 && x+dx < width && y+dy >= 0 && y+dy < height {
					targetIdx := (y+dy)*width + (x + dx)
					errR[targetIdx] += er / 8.0
					errG[targetIdx] += eg / 8.0
					errB[targetIdx] += eb / 8.0
				}
			}

			//  *   1   2
			// -1   0   1
			//  0   1
			//   *   1/8 1/8
			// 1/8 1/8 1/8
			//     1/8

			// dx=1, dy=0
			distributeError(1, 0, quantErrorR, quantErrorG, quantErrorB)
			// dx=2, dy=0
			distributeError(2, 0, quantErrorR, quantErrorG, quantErrorB)
			// dx=-1, dy=1
			distributeError(-1, 1, quantErrorR, quantErrorG, quantErrorB)
			// dx=0, dy=1
			distributeError(0, 1, quantErrorR, quantErrorG, quantErrorB)
			// dx=1, dy=1
			distributeError(1, 1, quantErrorR, quantErrorG, quantErrorB)
			// dx=0, dy=2
			distributeError(0, 2, quantErrorR, quantErrorG, quantErrorB)
		}
	}
}
