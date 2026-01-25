package image

import (
	"image"
	"image/color"
	"math"

	"github.com/mjibson/go-dsp/fft"
)

const (
	FreqThreshold     = 0.30
	TargetAngle       = 135.0
	AngleTolerance    = 10.0
	AttenuationFactor = 0.10
)

func (ip *ImageProcessor) RemoveRainbowEffect() {
	bounds := (*ip.Img).Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// 2. Arrays para almacenar los canales Y, Cb y Cr
	yChannel := make([][]float64, h)
	cbChannel := make([][]uint8, h)
	crChannel := make([][]uint8, h)

	for y := 0; y < h; y++ {
		yChannel[y] = make([]float64, w)
		cbChannel[y] = make([]uint8, w)
		crChannel[y] = make([]uint8, w)
		for x := 0; x < w; x++ {
			// Convertimos cada píxel a YCbCr
			r, g, b, _ := (*ip.Img).At(x, y).RGBA()
			yy, cb, cr := color.RGBToYCbCr(uint8(r>>8), uint8(g>>8), uint8(b>>8))

			yChannel[y][x] = float64(yy) // Solo este irá a la FFT
			cbChannel[y][x] = cb         // Guardamos color intacto
			crChannel[y][x] = cr         // Guardamos color intacto
		}
	}

	ip.Img = nil

	// 3. FFT 2D solo sobre el canal Y (Luminancia)
	coeffs := fft.FFT2Real(yChannel)

	// 4. Filtrado de frecuencias (Eliminación del patrón arcoíris)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Centrar frecuencias para el cálculo
			fy := float64(y)/float64(h) - 0.5
			fx := float64(x)/float64(w) - 0.5
			dist := math.Sqrt(fx*fx + fy*fy)

			if dist >= FreqThreshold {
				angle := math.Mod(math.Atan2(fy, fx)*180/math.Pi+360, 360)
				if isTargetAngle(angle) {
					// Aplicamos la atenuación al número complejo
					coeffs[y][x] = complex(real(coeffs[y][x])*AttenuationFactor, imag(coeffs[y][x])*AttenuationFactor)
				}
			}
		}
	}

	// 5. Inversa de FFT para recuperar la luminancia limpia
	cleanY := fft.IFFT2(coeffs)

	// 6. Reconstruir la imagen final combinando Y limpia + Cb/Cr originales
	outImg := image.NewRGBA(bounds)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			yVal := real(cleanY[y][x])
			if yVal > 255 {
				yVal = 255
			}
			if yVal < 0 {
				yVal = 0
			}

			// Combinamos la Y procesada con Cb y Cr que guardamos antes
			rgba := color.YCbCr{
				Y:  uint8(yVal),
				Cb: cbChannel[y][x],
				Cr: crChannel[y][x],
			}
			outImg.Set(x, y, rgba)
		}
	}

	outImgAsImg := image.Image(outImg)
	ip.Img = &outImgAsImg
}

func isTargetAngle(angle float64) bool {
	targets := []float64{135, 315, 225, 45}
	for _, t := range targets {
		diff := math.Abs(angle - t)
		if diff <= AngleTolerance || diff >= 360-AngleTolerance {
			return true
		}
	}
	return false
}
