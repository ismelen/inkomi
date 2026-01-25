package image

import (
	"image"
	"image/color"
)

type bBox struct {
	left, top, right, bottom int
}

func (b *bBox) getSurface() float64 {
	return float64(b.right - b.left) / float64(b.top - b.bottom)
}

type side struct {
	vector       int
	isHorizontal bool
	max          int
	origin       int
	end          int
}

const threshold = uint8(128)

func (ip *ImageProcessor) GetBBox(img image.Image, checkWhite bool) bBox {
	dim := img.Bounds()
	w, h := dim.Dx(), dim.Dy()
	maxX := int(0.1 * float64(w))
	maxY := int(0.1 * float64(h))
	checkWhite = !checkWhite

	top := side{
		vector:       1,
		isHorizontal: false,
		max:          maxY,
		end:          w,
		origin:       0,
	}

	bottom := side{
		vector:       -1,
		isHorizontal: false,
		max:          maxY,
		end:          w,
		origin:       h,
	}

	left := side{
		vector:       1,
		isHorizontal: true,
		max:          maxX,
		end:          h,
		origin:       0,
	}

	right := side{
		vector:       -1,
		isHorizontal: true,
		max:          maxX,
		end:          h,
		origin:       w,
	}

	return bBox{
		top:    top.getSide(img, checkWhite),
		bottom: bottom.getSide(img, checkWhite),
		left:   left.getSide(img, checkWhite),
		right:  right.getSide(img, checkWhite),
	}
}

func (ip *ImageProcessor) CalculateBboxAgresive(img image.Image, checkWhite bool) bBox {
	//TODO: Implement, without page number
	return ip.GetBBox(img, checkWhite)
}

func (this *side) getSide(img image.Image, checkWhite bool) int {

	var c color.Gray
	for lvl := 1; lvl != this.max; lvl++ {
		level := this.origin + lvl*this.vector
		for p := 1; p < this.end; p++ {
			if this.isHorizontal {
				c = color.GrayModel.Convert(img.At(level, p)).(color.Gray)
			} else {
				c = color.GrayModel.Convert(img.At(p, level)).(color.Gray)
			}
			if (checkWhite && c.Y >= threshold) ||
				(!checkWhite && c.Y < threshold) {
				return level - 1
			}
		}
	}

	return this.origin
}
