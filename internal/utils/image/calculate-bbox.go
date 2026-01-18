package ImageUtils

import (
	"image"
	"image/color"
)

type BBox struct {
	Left, Top, Right, Bottom int
}

const threshold = uint8(128)

type side struct {
	vector int
	isHorizontal bool
	max int
	fin int
	origin int
}

func (this *side) getSide(img image.Image, checkWhite bool) int {
	
	var c color.Gray
	for lvl:=1; lvl!=this.max; lvl++ {
		level := this.origin + lvl * this.vector
		for p:=1; p<this.fin; p++ {
			if this.isHorizontal {
				c = color.GrayModel.Convert(img.At(level, p)).(color.Gray)
				}else {
					c = color.GrayModel.Convert(img.At(p, level)).(color.Gray)
				}
				if (checkWhite && c.Y >= threshold) || 
				(!checkWhite && c.Y < threshold) {
				return level-1
			}
		}
	}

	return this.origin
}

func CalculateBbox(img image.Image, hasWhiteBg bool) BBox {
	dim := img.Bounds()
	w, h := dim.Dx(), dim.Dy()
	maxX := int(0.1 * float64(w))
	maxY := int(0.1 * float64(h))
	checkWhite := !hasWhiteBg


	top := side{
		vector: 1,
		isHorizontal: false,
		max: maxY,
		fin: w,
		origin: 0,
	}

	bottom := side{
		vector: -1,
		isHorizontal: false,
		max: maxY,
		fin: w,
		origin: h,
	}

	left := side{
		vector: 1,
		isHorizontal: true,
		max: maxX,
		fin: h,
		origin: 0,
	}

	right := side{
		vector: -1,
		isHorizontal: true,
		max: maxX,
		fin: h,
		origin: w,
	}

	return BBox{
		Top: top.getSide(img, checkWhite),
		Bottom: bottom.getSide(img, checkWhite),
		Left: left.getSide(img, checkWhite),
		Right: right.getSide(img, checkWhite),
	}
}


func CalculateBboxAgresive(img image.Image, hasWhiteBg bool) BBox {
	return CalculateBbox(img, hasWhiteBg)
}
