package ImageUtils

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"math"

	"github.com/mjibson/go-dsp/fft"
)

// Parámetros de filtrado (similares a tu script de Python)
const (
	FreqThreshold     = 0.30
	TargetAngle       = 135.0
	AngleTolerance    = 10.0
	AttenuationFactor = 0.10
)

func EraseRainbowArtifacts(img *image.Image) {
	bounds := (*img).Bounds()
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
			r, g, b, _ := (*img).At(x, y).RGBA()
			yy, cb, cr := color.RGBToYCbCr(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			
			yChannel[y][x] = float64(yy) // Solo este irá a la FFT
			cbChannel[y][x] = cb         // Guardamos color intacto
			crChannel[y][x] = cr         // Guardamos color intacto
		}
	}

	img = nil

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
			if yVal > 255 { yVal = 255 }
			if yVal < 0 { yVal = 0 }

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
	img = &outImgAsImg
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

// package ImageUtils

// import (
// 	"image"
// 	"image/color"
// 	"math"
// 	"math/cmplx"

// 	"github.com/disintegration/imaging"
// 	"github.com/mjibson/go-dsp/fft"
// )

// // Constantes para el filtro
// const (
// 	FreqThreshold     = 0.30
// 	TargetAngle       = 135.0
// 	AngleTolerance    = 10.0
// 	AttenuationFactor = 0.10
// )

// // EraseRainbowArtifacts es la función principal (Wrapper)
// func EraseRainbowArtifacts(img *image.Image, isColor bool) {
// 	if isColor {
// 		// 1. Convertir a NRGBA (para manipulación directa de píxeles)
// 		src := imaging.Clone((*img))
// 		bounds := src.Bounds()
// 		w, h := bounds.Dx(), bounds.Dy()

// 		// 2. RGB -> YUV (Trabajamos sobre la Luminancia Y)
// 		yValues := make([]float64, w*h)
// 		uValues := make([]float64, w*h)
// 		vValues := make([]float64, w*h)

// 		for y := 0; y < h; y++ {
// 			for x := 0; x < w; x++ {
// 				r, g, b, _ := src.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
// 				R, G, B := float64(r>>8), float64(g>>8), float64(b>>8)

// 				// Matriz de conversión standard
// 				Y := 0.299*R + 0.587*G + 0.114*B
// 				U := -0.14713*R - 0.28886*G + 0.436*B
// 				V := 0.615*R - 0.51499*G - 0.10001*B

// 				yValues[y*w+x] = Y
// 				uValues[y*w+x] = U
// 				vValues[y*w+x] = V
// 			}
// 		}

// 		// 3. FFT 2D sobre canal Y
// 		fftResult := FourierTransform2D(yValues, w, h)

// 		// 4. Atenuar frecuencias diagonales
// 		AttenuateDiagonalFrequencies(fftResult, w, h)

// 		// 5. IFFT 2D
// 		cleanY := InverseFourierTransform2D(fftResult, w, h)

// 		// 6. YUV -> RGB y Reconstruir imagen
// 		dst := image.NewRGBA(image.Rect(0, 0, w, h))
// 		for y := 0; y < h; y++ {
// 			for x := 0; x < w; x++ {
// 				Y := cleanY[y*w+x]
// 				U := uValues[y*w+x]
// 				V := vValues[y*w+x]

// 				R := uint8(math.Max(0, math.Min(255, Y+1.13983*V)))
// 				G := uint8(math.Max(0, math.Min(255, Y-0.39465*U-0.58060*V)))
// 				B := uint8(math.Max(0, math.Min(255, Y+2.03211*U)))

// 				dst.Set(x, y, color.RGBA{R, G, B, 255})
// 			}
// 		}
// 		(*img) = dst
// 	} else {
// 		// Procesamiento Grayscale simple
// 		gray := imaging.Grayscale((*img))
// 		bounds := gray.Bounds()
// 		w, h := bounds.Dx(), bounds.Dy()
// 		pixels := make([]float64, w*h)
// 		for y := 0; y < h; y++ {
// 			for x := 0; x < w; x++ {
// 				r, _, _, _ := gray.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
// 				pixels[y*w+x] = float64(r >> 8)
// 			}
// 		}
// 		fftResult := FourierTransform2D(pixels, w, h)
// 		AttenuateDiagonalFrequencies(fftResult, w, h)
// 		cleanPixels := InverseFourierTransform2D(fftResult, w, h)

// 		dst := image.NewGray(image.Rect(0, 0, w, h))
// 		for y := 0; y < h; y++ {
// 			for x := 0; x < w; x++ {
// 				dst.SetGray(x, y, color.Gray{uint8(math.Max(0, math.Min(255, cleanPixels[y*w+x])))})
// 			}
// 		}
// 		(*img) = dst
// 	}
// }

// // FourierTransform2D realiza una FFT de filas y luego de columnas
// func FourierTransform2D(data []float64, w, h int) [][]complex128 {
// 	// FFT en filas
// 	rowsFFT := make([][]complex128, h)
// 	for y := 0; y < h; y++ {
// 		row := data[y*w : (y+1)*w]
// 		rowsFFT[y] = fft.FFTReal(row)
// 	}

// 	// Transponer y FFT en columnas
// 	result := make([][]complex128, w)
// 	for x := 0; x < w; x++ {
// 		col := make([]complex128, h)
// 		for y := 0; y < h; y++ {
// 			col[y] = rowsFFT[y][x]
// 		}
// 		result[x] = fft.FFT(col)
// 	}
// 	return result // Resultado transpuesto [x][y]
// }

// // InverseFourierTransform2D realiza la IFFT inversa
// func InverseFourierTransform2D(data [][]complex128, w, h int) []float64 {
// 	colsIFFT := make([][]complex128, w)
// 	for x := 0; x < w; x++ {
// 		colsIFFT[x] = fft.IFFT(data[x])
// 	}

// 	finalData := make([]float64, w*h)
// 	for y := 0; y < h; y++ {
// 		row := make([]complex128, w)
// 		for x := 0; x < w; x++ {
// 			row[x] = colsIFFT[x][y]
// 		}
// 		rowInverse := fft.IFFT(row)
// 		for x := 0; x < w; x++ {
// 			finalData[y*w+x] = real(rowInverse[x])
// 		}
// 	}
// 	return finalData
// }

// // AttenuateDiagonalFrequencies aplica el filtro en el dominio de Fourier
// func AttenuateDiagonalFrequencies(spectrum [][]complex128, w, h int) {
// 	wHalf := len(spectrum)
// 	hHalf := len(spectrum[0])

// 	targetAngles := []float64{
// 		TargetAngle,
// 		math.Mod(TargetAngle+180, 360),
// 		math.Mod(TargetAngle+90, 360),
// 		math.Mod(TargetAngle+270, 360),
// 	}

// 	for x := 0; x < wHalf; x++ {
// 		for y := 0; y < hHalf; y++ {
// 			// Calcular frecuencias normalizadas
// 			fx := float64(x) / float64(w)
// 			fy := float64(y) / float64(h)
// 			if y > hHalf/2 {
// 				fy -= 1.0
// 			} // Ajuste para frecuencias negativas

// 			distSq := fx*fx + fy*fy
// 			if distSq >= FreqThreshold*FreqThreshold {
// 				angle := math.Atan2(fy, fx) * 180 / math.Pi
// 				if angle < 0 {
// 					angle += 360
// 				}

// 				for _, target := range targetAngles {
// 					diff := math.Abs(angle - target)
// 					if diff > 180 {
// 						diff = 360 - diff
// 					}

// 					if diff <= AngleTolerance {
// 						spectrum[x][y] = cmplx.Rect(cmplx.Abs(spectrum[x][y])*AttenuationFactor, cmplx.Phase(spectrum[x][y]))
// 						break
// 					}
// 				}
// 			}
// 		}
// 	}
// }
