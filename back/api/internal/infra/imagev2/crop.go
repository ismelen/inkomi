package imagev2

import (
	"image"
	"image/color"
)

// calculateCropBox scans the image from the 4 edges to the center
// to find the content bounding box (ink).
func calculateCropBox(img image.Image) image.Rectangle {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Threshold for darkness (0-255). Lower is darker.
	// Let's say if luminance is < 200, it's considered ink.
	// Or we can use binarization threshold (e.g., 200).
	const darkThreshold = 200

	// Tolerance: > 2% of pixels must be darker than threshold
	xTolerance := int(float64(height) * 0.02)
	yTolerance := int(float64(width) * 0.02)

	if xTolerance < 1 {
		xTolerance = 1
	}
	if yTolerance < 1 {
		yTolerance = 1
	}

	isDarkPixel := func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		// Convert to luminance (0-255)
		// Y = 0.299 R + 0.587 G + 0.114 B
		// Since RGBA() returns 16-bit pre-multiplied values (0-65535), we divide by 256
		y := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0
		return y < darkThreshold
	}

	minY := bounds.Min.Y
	maxY := bounds.Max.Y
	minX := bounds.Min.X
	maxX := bounds.Max.X

	// Scan from Top
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		darkCount := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isDarkPixel(img.At(x, y)) {
				darkCount++
			}
		}
		if darkCount > yTolerance {
			minY = y
			break
		}
	}

	// Scan from Bottom
	for y := bounds.Max.Y - 1; y >= minY; y-- {
		darkCount := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isDarkPixel(img.At(x, y)) {
				darkCount++
			}
		}
		if darkCount > yTolerance {
			maxY = y + 1
			break
		}
	}

	// Scan from Left
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		darkCount := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if isDarkPixel(img.At(x, y)) {
				darkCount++
			}
		}
		if darkCount > xTolerance {
			minX = x
			break
		}
	}

	// Scan from Right
	for x := bounds.Max.X - 1; x >= minX; x-- {
		darkCount := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if isDarkPixel(img.At(x, y)) {
				darkCount++
			}
		}
		if darkCount > xTolerance {
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
