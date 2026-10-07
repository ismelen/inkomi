package imagev2

import (
	"image"
	"math"
)

// calculateCropBox scans the image from the 4 edges to the center
// to find the content bounding box (ink). It dynamically adapts to
// black, white, or grey pages by calculating the baseline luminance
// of each margin and looking for contrasting pixels.
func calculateCropBox(img image.Image, tolerance float64) image.Rectangle {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Tolerance: percentage of the dimension that must differ from the margin baseline.
	// If <= 0, default to 2% of the dimension.
	if tolerance <= 0 {
		tolerance = 0.02
	} else if tolerance > 1.0 {
		// Just in case someone passes percentages like 2.0 instead of 0.02
		tolerance = tolerance / 100.0
	}

	xTolerance := int(float64(height) * tolerance)
	yTolerance := int(float64(width) * tolerance)

	if xTolerance < 1 { xTolerance = 1 }
	if yTolerance < 1 { yTolerance = 1 }

	// Minimum luminance difference to be considered "content" (out of 255)
	const contrastThreshold = 40.0

	getLuma := func(x, y int) float64 {
		r, g, b, _ := img.At(x, y).RGBA()
		return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0
	}

	getLineAvgLuma := func(isHorizontal bool, fixedIdx, minIdx, maxIdx int) float64 {
		sum := 0.0
		for i := minIdx; i < maxIdx; i++ {
			if isHorizontal {
				sum += getLuma(i, fixedIdx)
			} else {
				sum += getLuma(fixedIdx, i)
			}
		}
		return sum / float64(maxIdx-minIdx)
	}

	minY := bounds.Min.Y
	maxY := bounds.Max.Y
	minX := bounds.Min.X
	maxX := bounds.Max.X

	// 1. Scan from Top
	topBaseline := getLineAvgLuma(true, bounds.Min.Y, bounds.Min.X, bounds.Max.X)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		diffCount := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if math.Abs(getLuma(x, y)-topBaseline) > contrastThreshold {
				diffCount++
			}
		}
		if diffCount > yTolerance {
			minY = y
			break
		}
	}

	// 2. Scan from Bottom
	bottomBaseline := getLineAvgLuma(true, bounds.Max.Y-1, bounds.Min.X, bounds.Max.X)
	for y := bounds.Max.Y - 1; y >= minY; y-- {
		diffCount := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if math.Abs(getLuma(x, y)-bottomBaseline) > contrastThreshold {
				diffCount++
			}
		}
		if diffCount > yTolerance {
			maxY = y + 1
			break
		}
	}

	// 3. Scan from Left
	leftBaseline := getLineAvgLuma(false, bounds.Min.X, bounds.Min.Y, bounds.Max.Y)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		diffCount := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if math.Abs(getLuma(x, y)-leftBaseline) > contrastThreshold {
				diffCount++
			}
		}
		if diffCount > xTolerance {
			minX = x
			break
		}
	}

	// 4. Scan from Right
	rightBaseline := getLineAvgLuma(false, bounds.Max.X-1, bounds.Min.Y, bounds.Max.Y)
	for x := bounds.Max.X - 1; x >= minX; x-- {
		diffCount := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if math.Abs(getLuma(x, y)-rightBaseline) > contrastThreshold {
				diffCount++
			}
		}
		if diffCount > xTolerance {
			maxX = x + 1
			break
		}
	}

	// Safety check
	if minX >= maxX || minY >= maxY {
		return bounds
	}

	return image.Rect(minX, minY, maxX, maxY)
}
