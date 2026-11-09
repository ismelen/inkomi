package imaging

import (
	"image"
	"image/draw"
)

// ToRGBA returns img as an *image.RGBA starting at the origin, copying it only when needed.
func ToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok && rgba.Rect.Min == (image.Point{}) {
		return rgba
	}
	b := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Rect, img, b.Min, draw.Src)
	return rgba
}
