package ImageUtils

import (
	"image"
	"image/color"
)

type ColorDetector struct {
	ForceColor bool
}

func (d *ColorDetector) CalculateColor(img image.Image) bool {
	// 1. Obtener histogramas de Cb y Cr
	cbHist, crHist := getYCbCrHistograms(img)

	// Definición de las pasadas (cutoff, diffThreshold)
	// Nota: En tu Python usas floats (.2) y enteros (3). Aquí los manejamos como floats.
	steps := []struct {
		cutoff        float64
		diffThreshold int
	}{
		{0, 22},
		{0.2, 10},
		{3.0, 4},
	}

	for _, step := range steps {
		done, decision := d.colorPrecision(cbHist, crHist, step.cutoff, step.diffThreshold)
		if done {
			return decision
		}
	}

	return false
}

func getYCbCrHistograms(img image.Image) (cbHist [256]int, crHist [256]int) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			// Convertimos el color actual a YCbCr
			cycbcr := color.YCbCrModel.Convert(c).(color.YCbCr)
			cbHist[cycbcr.Cb]++
			crHist[cycbcr.Cr]++
		}
	}
	return
}

func (d *ColorDetector) colorPrecision(cbHistOrig, crHistOrig [256]int, cutoff float64, diffThreshold int) (bool, bool) {
	// Aplicar cutoff (limpieza de ruido en los extremos del histograma)
	cbHist := applyCutoff(cbHistOrig, cutoff)
	crHist := applyCutoff(crHistOrig, cutoff)

	// Obtener índices non-zero
	cbMin, cbMax, cbOk := getNonZeroRange(cbHist)
	crMin, crMax, crOk := getNonZeroRange(crHist)

	if !cbOk || !crOk {
		return true, false
	}

	cbSpread := cbMax - cbMin
	crSpread := crMax - crMin

	const SpreadThreshold = 7

	if d.ForceColor {
		if cbMin > 128 || crMin > 128 || cbMax < 128 || crMax < 128 {
			return true, true
		}
	} else if cbSpread < SpreadThreshold && crSpread < SpreadThreshold {
		return true, false
	}

	if cbMin <= 128-diffThreshold || crMin <= 128-diffThreshold ||
		cbMax >= 128+diffThreshold || crMax >= 128+diffThreshold {
		return true, true
	}

	return false, false
}

// Función auxiliar para limpiar el histograma (simula histograms_cutoff)
func applyCutoff(hist [256]int, cutoff float64) [256]int {
	if cutoff <= 0 {
		return hist
	}
	
	total := 0
	for _, v := range hist {
		total += v
	}

	// Calculamos cuántos píxeles debemos ignorar (por ejemplo 0.2% o 3 píxeles fijos)
	// Si cutoff es < 1, lo tratamos como porcentaje. Si es > 1, como cantidad fija.
	var threshold int
	if cutoff < 1.0 {
		threshold = int(float64(total) * (cutoff / 100.0))
	} else {
		threshold = int(cutoff)
	}

	// Limpiar inicio
	count := 0
	for i := 0; i < 256; i++ {
		if hist[i] > 0 {
			count += hist[i]
			if count <= threshold {
				hist[i] = 0
			} else {
				break
			}
		}
	}
	// Limpiar fin
	count = 0
	for i := 255; i >= 0; i-- {
		if hist[i] > 0 {
			count += hist[i]
			if count <= threshold {
				hist[i] = 0
			} else {
				break
			}
		}
	}
	return hist
}

func getNonZeroRange(hist [256]int) (int, int, bool) {
	min, max := -1, -1
	for i := 0; i < 256; i++ {
		if hist[i] > 0 {
			if min == -1 {
				min = i
			}
			max = i
		}
	}
	return min, max, min != -1
}