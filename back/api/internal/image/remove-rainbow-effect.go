package image

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
)

const BlurFactor = 2.0

func (ip *ImageProcessor) RemoveRainbowEffect() {
	bounds := (*ip.Img).Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	grayImg := imaging.Grayscale(*ip.Img)
	blurredImg := imaging.Blur(*ip.Img, BlurFactor)

	outImg := image.NewRGBA(bounds)
	
	for y := range h {
		for x := range w {
			gR, _, _, _ := grayImg.At(x, y).RGBA()
			yVal := uint8(gR >> 8)

			cR, cG, cB, _ := blurredImg.At(x, y).RGBA()
			_, cb, cr := color.RGBToYCbCr(uint8(cR>>8), uint8(cG >> 8), uint8(cB >> 8))

			color := color.YCbCr{
				Y: yVal,
				Cb: cb,
				Cr: cr,
			}
			outImg.Set(x, y, color)
		}
	}

	*ip.Img = outImg
}
