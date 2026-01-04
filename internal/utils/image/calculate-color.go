package ImageUtils

import (
	"image"
	"math"
)

// CalculateColor determina si la imagen debe tratarse como color o blanco y negro.
func CalculateColor(colorMode bool, img *image.Image) bool {
	// 1. Obtener histogramas de Cb y Cr
	cbHistOriginal, crHistOriginal := getChromaHistograms(img)

	// Definición de los pasos de precisión (cutoff_low, cutoff_high, threshold)
	steps := []struct {
		cutoffLow     float64
		cutoffHigh    float64
		diffThreshold int
	}{
		{0, 0, 22},
		{0.2, 0.2, 10},
		{3, 3, 4},
	}

	for _, step := range steps {
		done, decision := colorPrecision(cbHistOriginal, crHistOriginal, step.cutoffLow, step.cutoffHigh, step.diffThreshold, colorMode)
		if done {
			return decision
		}
	}

	return false
}

func colorPrecision(cbOrig, crOrig []int, cutLow, cutHigh float64, diffThreshold int, colorMode bool) (bool, bool) {
	// Copiamos los histogramas para no alterar los originales en cada iteración
	cbHist := make([]int, 256)
	crHist := make([]int, 256)
	copy(cbHist, cbOrig)
	copy(crHist, crOrig)

	// Aplicar el recorte de ruido (cutoff)
	cbHist, crHist = histogramsCutoff(cbHist, crHist, cutLow, cutHigh)

	// Encontrar índices no nulos (mínimo y máximo)
	cbMin, cbMax := findNonZeroRange(cbHist)
	crMin, crMax := findNonZeroRange(crHist)

	// Si no hay píxeles (histograma vacío), asumimos que no hay color
	if cbMin == -1 || crMin == -1 {
		return true, false
	}

	cbSpread := cbMax - cbMin
	crSpread := crMax - crMin

	const SPREAD_THRESHOLD = 7

	// Lógica de ForceColor
	if colorMode {
		if cbMin > 128 || crMin > 128 || cbMax < 128 || crMax < 128 {
			return true, true
		}
	} else if cbSpread < SPREAD_THRESHOLD && crSpread < SPREAD_THRESHOLD {
		// Si la dispersión es muy pequeña alrededor del centro, es B/N
		return true, false
	}

	// Comprobar si algún píxel se aleja del centro (128) más allá del umbral
	if cbMin <= 128-diffThreshold || crMin <= 128-diffThreshold ||
		cbMax >= 128+diffThreshold || crMax >= 128+diffThreshold {
		return true, true
	}

	return false, false
}

func histogramsCutoff(cbHist, crHist []int, cutLow, cutHigh float64) ([]int, []int) {
	if cutLow == 0 && cutHigh == 0 {
		return cbHist, crHist
	}

	hists := [][]int{cbHist, crHist}
	for _, h := range hists {
		totalPixels := 0
		for _, count := range h {
			totalPixels += count
		}

		// Recorte del extremo bajo
		cutL := int(float64(totalPixels) * cutLow / 100.0)
		for i := 0; i < 256 && cutL > 0; i++ {
			if cutL >= h[i] {
				cutL -= h[i]
				h[i] = 0
			} else {
				h[i] -= cutL
				cutL = 0
			}
		}

		// Recorte del extremo alto
		cutH := int(float64(totalPixels) * cutHigh / 100.0)
		for i := 255; i >= 0 && cutH > 0; i-- {
			if cutH >= h[i] {
				cutH -= h[i]
				h[i] = 0
			} else {
				h[i] -= cutH
				cutH = 0
			}
		}
	}
	return cbHist, crHist
}

// Helper para calcular histogramas de croma en una sola pasada
func getChromaHistograms(img *image.Image) ([]int, []int) {
	cbHist := make([]int, 256)
	crHist := make([]int, 256)
	bounds := (*img).Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := (*img).At(x, y).RGBA()
			// Convertir a 8-bit
			R, G, B := float64(r>>8), float64(g>>8), float64(b>>8)

			// Fórmulas estándar de conversión a YCbCr (puntos Cb y Cr)
			cb := 128 - 0.168736*R - 0.331264*G + 0.5*B
			cr := 128 + 0.5*R - 0.418688*G - 0.081312*B

			cbHist[clampUint8(cb)]++
			crHist[clampUint8(cr)]++
		}
	}
	return cbHist, crHist
}

func findNonZeroRange(hist []int) (int, int) {
	min, max := -1, -1
	for i := 0; i < 256; i++ {
		if hist[i] > 0 {
			min = i
			break
		}
	}
	if min == -1 {
		return -1, -1
	}
	for i := 255; i >= 0; i-- {
		if hist[i] > 0 {
			max = i
			break
		}
	}
	return min, max
}

func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}
