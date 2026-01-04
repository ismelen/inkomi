package ImageUtils

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

// --- Configuración de constantes (Porcentajes) ---
var (
	maxShapeSizeToleratedSize = [2]float64{0.045, 0.02} // 0.015*3, 0.02
	minShapeSizeToleratedSize = [2]float64{0.003, 0.006}
	windowHSize               = 0.02 * 1.25
	maxDistSize               = [2]float64{0.01, 0.002}
)

// BBox representa el bounding box (left, top, right, bottom)
type BBox struct {
	MinX, MinY, MaxX, MaxY int
}

// thresholdFromPower es un placeholder para la lógica de common_crop.py
// Ajusta esta fórmula según tus necesidades reales.
func thresholdFromPower(power float64) uint8 {
	// Ejemplo: a mayor power, threshold más alto (más agresivo)
	return uint8(math.Max(0, math.Min(255, 128+(power*20))))
}

// getBBox busca los límites de los píxeles no negros (equivalente a PIL.getbbox)
func getBBox(img image.Image) *BBox {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := -1, -1
	found := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Convertimos a Gray para evaluar intensidad
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y > 0 { // Si no es negro
				found = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if !found {
		return nil
	}
	return &BBox{minX, minY, maxX, maxY}
}

// GetBBoxCropMarginPageNumber intenta detectar y eliminar el número de página inferior
func GetBBoxCropMarginPageNumber(img image.Image, power float64, backgroundColor string) *BBox {
	// 1. Grayscale e Invertir si es necesario
	grayImg := imaging.Grayscale(img)
	var processed image.Image = grayImg
	if backgroundColor != "white" {
		processed = imaging.Invert(processed)
	}

	// 2. Autocontrast (Simulado) y Blur
	// imaging no tiene Autocontrast idéntico a PIL, usamos AdjustContrast
	processed = imaging.AdjustContrast(processed, 10)
	processed = imaging.Blur(processed, 1.0)

	// 3. Crear imagen Blanco y Negro (Threshold)
	threshold := thresholdFromPower(power)
	bounds := processed.Bounds()
	bwImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := color.GrayModel.Convert(processed.At(x, y)).(color.Gray).Y
			if p <= threshold {
				bwImg.SetGray(x, y, color.Gray{Y: 255}) // "Negro" en el original ahora es Blanco
			} else {
				bwImg.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	IgnorePixelsNearEdge(bwImg)
	globalBBox := getBBox(bwImg)
	if globalBBox == nil {
		return nil
	}

	// 4. Analizar ventana inferior
	imgW, imgH := bounds.Dx(), bounds.Dy()
	windowH := int(float64(imgH) * windowHSize)

	// Crop de la parte inferior sospechosa
	rectPart := image.Rect(globalBBox.MinX, globalBBox.MaxY-windowH, globalBBox.MaxX, globalBBox.MaxY)
	imgPart := bwImg.SubImage(rectPart).(*image.Gray)

	// 5. Detectar grupos de píxeles (equivalente a np.where + group_close_values)
	var windowGroups []BBox
	distX := int(float64(imgW) * maxDistSize[0])
	partBounds := imgPart.Bounds()

	for y := partBounds.Min.Y; y < partBounds.Max.Y; y++ {
		var activeX []int
		for x := partBounds.Min.X; x < partBounds.Max.X; x++ {
			if imgPart.GrayAt(x, y).Y == 255 {
				activeX = append(activeX, x)
			}
		}
		if len(activeX) > 0 {
			groups := GroupCloseValues(activeX, distX)
			for _, g := range groups {
				windowGroups = append(windowGroups, BBox{g[0], y, g[1], y})
			}
		}
	}

	// 6. Mezclar cajas cercanas
	distY := int(float64(imgH) * maxDistSize[1])
	mergedBoxes := MergeBoxes(windowGroups, distX, distY)

	// 7. Filtrar por tamaño y posición
	minW := int(float64(imgW) * minShapeSizeToleratedSize[0])
	minH := int(float64(imgH) * minShapeSizeToleratedSize[1])

	var validBoxes []BBox
	var lowestBoxes []BBox
	for _, b := range mergedBoxes {
		if (b.MaxX-b.MinX >= minW) && (b.MaxY-b.MinY >= minH) {
			validBoxes = append(validBoxes, b)
			if b.MaxY >= partBounds.Max.Y-1 {
				lowestBoxes = append(lowestBoxes, b)
			}
		}
	}

	finalBotY := globalBBox.MaxY
	if len(lowestBoxes) > 0 {
		minYOfLowest := lowestBoxes[0].MinY
		for _, b := range lowestBoxes {
			if b.MinY < minYOfLowest {
				minYOfLowest = b.MinY
			}
		}

		var sameYRange []BBox
		for _, b := range validBoxes {
			if b.MaxY >= minYOfLowest {
				sameYRange = append(sameYRange, b)
			}
		}

		maxWTol := int(float64(imgW) * maxShapeSizeToleratedSize[0])
		maxHTol := int(math.Max(float64(imgH)*maxShapeSizeToleratedSize[1], 3))

		if len(sameYRange) == 1 {
			b := sameYRange[0]
			if (b.MaxX-b.MinX <= maxWTol) && (b.MaxY-b.MinY <= maxHTol) {
				// Calculamos el punto de corte
				finalBotY = globalBBox.MaxY - (windowH - b.MinY + 1)
			}
		}
	}

	// 8. Recorte final y obtener BBox resultante
	finalRect := image.Rect(0, 0, imgW, finalBotY)
	croppedFinal := imaging.Crop(bwImg, finalRect)
	return getBBox(croppedFinal)
}

// IgnorePixelsNearEdge limpia ruido en los bordes extremos (2%)
func IgnorePixelsNearEdge(bwImg *image.Gray) {
	w, h := bwImg.Bounds().Dx(), bwImg.Bounds().Dy()
	edgeRects := []image.Rectangle{
		image.Rect(0, 0, w, int(0.02*float64(h))),
		image.Rect(0, int(0.98*float64(h)), w, h),
		image.Rect(0, 0, int(0.02*float64(w)), h),
		image.Rect(int(0.98*float64(w)), 0, w, h),
	}

	for _, r := range edgeRects {
		whitePixels := 0
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				if bwImg.GrayAt(x, y).Y == 255 {
					whitePixels++
				}
			}
		}

		imperfections := float64(whitePixels) / float64(r.Dx()*r.Dy())
		if imperfections > 0 && imperfections < 0.02 {
			for y := r.Min.Y; y < r.Max.Y; y++ {
				for x := r.Min.X; x < r.Max.X; x++ {
					bwImg.SetGray(x, y, color.Gray{Y: 0})
				}
			}
		}
	}
}

// GroupCloseValues agrupa coordenadas X que están a menos de maxDist
func GroupCloseValues(arr []int, maxDist int) [][2]int {
	if len(arr) == 0 {
		return nil
	}
	var groups [][2]int
	start := arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i]-arr[i-1] > maxDist {
			groups = append(groups, [2]int{start, arr[i-1]})
			start = arr[i]
		}
	}
	groups = append(groups, [2]int{start, arr[len(arr)-1]})
	return groups
}

// MergeBoxes une rectángulos que se solapan o están muy cerca
func MergeBoxes(boxes []BBox, distX, distY int) []BBox {
	if len(boxes) == 0 {
		return nil
	}

	for {
		changed := false
		for i := 0; i < len(boxes); i++ {
			for j := i + 1; j < len(boxes); j++ {
				if intersects(boxes[i], boxes[j], distX, distY) {
					// Expandir i para incluir j
					boxes[i].MinX = minInt(boxes[i].MinX, boxes[j].MinX)
					boxes[i].MaxX = maxInt(boxes[i].MaxX, boxes[j].MaxX)
					boxes[i].MinY = minInt(boxes[i].MinY, boxes[j].MinY)
					boxes[i].MaxY = maxInt(boxes[i].MaxY, boxes[j].MaxY)

					// Eliminar j
					boxes = append(boxes[:j], boxes[j+1:]...)
					changed = true
					break
				}
			}
			if changed {
				break
			}
		}
		if !changed {
			break
		}
	}
	return boxes
}

func intersects(a, b BBox, dx, dy int) bool {
	return !(b.MinX-dx > a.MaxX || b.MaxX+dx < a.MinX || b.MinY-dy > a.MaxY || b.MaxY+dy < a.MinY)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Aquí llamarías a GetBBoxCropMarginPageNumber con tu imagen cargada
}
