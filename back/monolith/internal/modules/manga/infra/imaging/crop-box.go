package imaging

import (
	"image"
	"image/draw"
	"math"
)

// contrastThreshold is the luma difference (out of 255) from a margin baseline that counts as ink.
const contrastThreshold = 40.0

// CropBox scans img from its 4 edges towards the center and returns the content bounding box.
// Each margin's baseline is its own average luma, so it adapts to white, black and grey pages.
// tolerance is the fraction of a line that must differ from the baseline to count as content.
func CropBox(img image.Image, tolerance float64) image.Rectangle {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return bounds
	}
	if tolerance <= 0 {
		tolerance = 0.02
	}

	rowTolerance := max(1, int(float64(width)*tolerance))
	colTolerance := max(1, int(float64(height)*tolerance))
	luma := lumaOf(img)
	at := func(x, y int) float64 { return float64(luma.Pix[y*luma.Stride+x]) }

	rowInk := func(y int, baseline float64) int {
		n := 0
		for x := range width {
			if math.Abs(at(x, y)-baseline) > contrastThreshold {
				n++
			}
		}
		return n
	}
	colInk := func(x int, baseline float64) int {
		n := 0
		for y := range height {
			if math.Abs(at(x, y)-baseline) > contrastThreshold {
				n++
			}
		}
		return n
	}
	rowAvg := func(y int) float64 {
		sum := 0.0
		for x := range width {
			sum += at(x, y)
		}
		return sum / float64(width)
	}
	colAvg := func(x int) float64 {
		sum := 0.0
		for y := range height {
			sum += at(x, y)
		}
		return sum / float64(height)
	}

	minY, maxY, minX, maxX := 0, height, 0, width

	top := rowAvg(0)
	for y := range height {
		if rowInk(y, top) > rowTolerance {
			minY = y
			break
		}
	}
	bottom := rowAvg(height - 1)
	for y := height - 1; y >= minY; y-- {
		if rowInk(y, bottom) > rowTolerance {
			maxY = y + 1
			break
		}
	}
	left := colAvg(0)
	for x := range width {
		if colInk(x, left) > colTolerance {
			minX = x
			break
		}
	}
	right := colAvg(width - 1)
	for x := width - 1; x >= minX; x-- {
		if colInk(x, right) > colTolerance {
			maxX = x + 1
			break
		}
	}

	if minX >= maxX || minY >= maxY {
		return bounds
	}
	return image.Rect(minX, minY, maxX, maxY).Add(bounds.Min)
}

func lumaOf(img image.Image) *image.Gray {
	if gray, ok := img.(*image.Gray); ok && gray.Rect.Min == (image.Point{}) {
		return gray
	}
	b := img.Bounds()
	gray := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(gray, gray.Rect, img, b.Min, draw.Src)
	return gray
}
