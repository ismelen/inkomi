package ImageUtils

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

func FillCheck(width, height int, img *image.Image) string {
	bw := imaging.Grayscale(*img)
	boxA, err1 := getBBoxInternal(bw, true)
	boxB, err2 := getBBoxInternal(bw, false)

	var surfaceB, surfaceW float64
	var diff float64

	if err1 != nil || err2 != nil {
		surfaceB, surfaceW = 0, 0
		diff = 0
	} else {
		surfaceB = float64((boxB.X.Max - boxB.X.Min) * (boxB.Y.Max - boxB.Y.Min)) // Box of blacks
		surfaceW = float64((boxA.X.Max - boxA.X.Min) * (boxA.Y.Max - boxA.Y.Min)) // Box of whites
		minSurf := math.Min(surfaceB, surfaceW)
		if minSurf == 0 {
			diff = 0 // Avoid divide by zero
		} else {
			diff = ((math.Max(surfaceB, surfaceW) - minSurf) / minSurf) * 100
		}
	}

	if diff > 0.5 {
		if surfaceW < surfaceB {
			return "white"
		} else if surfaceW > surfaceB {
			return "black"
		}
	}

	return "white"
}

type bBox struct {
	X MinMax
	Y MinMax
}

type MinMax struct {
	Min int
	Max int
}

func getBBoxInternal(img image.Image, checkForWhite bool) (*bBox, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	box := bBox{
		X: MinMax{
			Min: width,
			Max: -1,
		},
		Y: MinMax{
			Min: height,
			Max: -1,
		},
	}

	found := false

	threshold := uint8(128)

	// Early exit: check if we need to scan at all
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			isHit := (checkForWhite && c.Y >= threshold) || (!checkForWhite && c.Y < threshold)

			if isHit {
				found = true
				if x < box.X.Min {
					box.X.Min = x
				}
				if x > box.X.Max {
					box.X.Max = x
				}
				if y < box.Y.Min {
					box.Y.Min = y
				}
				if y > box.Y.Max {
					box.Y.Max = y
				}
			}
		}
	}

	if found {
		box.X.Max++
		box.Y.Max++
		return &box, nil
	}
	return nil, fmt.Errorf("cannot calculate bbox")
}
