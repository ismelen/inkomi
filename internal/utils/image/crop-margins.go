package ImageUtils

import (
	"image"
	manga "ismelen/ermc/internal/manga/logic/models"
	"math"

	"github.com/disintegration/imaging"
)

func CropMargins(
	payload *manga.PagePayload,
	power float64,
	fill string,
	minimum float64,
	preserveMargin float64,
) {
	bbox := GetBBoxCropMarginPageNumber(
		*payload.Image,
		power,
		fill,
	)

	if bbox == nil {
		return
	}

	bounds := (*payload.Image).Bounds()
	w, h := float64(bounds.Dx()), float64(bounds.Dy())

	left := math.Min(0.1*w, float64(bbox.MinX))
	upper := math.Min(0.1*h, float64(bbox.MinY))
	right := math.Max(0.9*w, float64(bbox.MaxX))
	lower := math.Max(0.9*h, float64(bbox.MaxY))

	// Convertimos de nuevo a nuestro objeto BBox o image.Rectangle
	refinedBBox := BBox{
		MinX: int(left),
		MinY: int(upper),
		MaxX: int(right),
		MaxY: int(lower),
	}

	// 3. Intentar el recorte
	MaybeCrop(payload.Image, refinedBBox, minimum, preserveMargin)
}

func MaybeCrop(img *image.Image, box BBox, minimum, preserveMargin float64) {
	bounds := (*img).Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	// Convertimos los límites a float64 para cálculos de precisión
	left := float64(box.MinX)
	upper := float64(box.MinY)
	right := float64(box.MaxX)
	lower := float64(box.MaxY)

	// 1. Lógica de preservar márgenes
	// Si preservemargin es 5 (5%), el ratio es 0.95.
	// Esto expande el área de recorte hacia afuera de nuevo.
	if preserveMargin > 0 {
		ratio := 1.0 - float64(preserveMargin)/100.0

		left = left * ratio
		upper = upper * ratio
		right = right + (w-right)*(1.0-ratio)
		lower = lower + (h-lower)*(1.0-ratio)
	}

	// 2. Cálculo de áreas para validar el recorte
	boxArea := (right - left) * (lower - upper)
	imageArea := w * h

	// Solo aplicamos el crop si el área resultante respeta el mínimo
	if (boxArea / imageArea) >= minimum {
		rect := image.Rect(
			int(math.Round(left)),
			int(math.Round(upper)),
			int(math.Round(right)),
			int(math.Round(lower)),
		)

		// imaging.Crop devuelve una nueva imagen recortada
		(*img) = imaging.Crop(*img, rect)
	}
}

func CropPageNumber(
	payload *manga.PagePayload,
	power float64,
	fill string,
	minimum float64,
	preserveMargin float64,
) {
	bbox := GetBBoxCropMarginPageNumber(
		*payload.Image,
		power,
		fill,
	)

	if bbox != nil {
		bounds := (*payload.Image).Bounds()
		w := float64(bounds.Dx())
		h := float64(bounds.Dy())

		// 2. Extraer coordenadas del BBox detectado
		left := float64(bbox.MinX)
		upper := float64(bbox.MinY)
		right := float64(bbox.MaxX)
		lower := float64(bbox.MaxY)

		// 3. Aplicar restricción de seguridad (No recortar más del 10%)
		// Python: bbox = (min(0.1*w, left), min(0.1*h, upper), max(0.9*w, right), max(0.9*h, lower))
		// Esto asegura que el "rectángulo de contenido" resultante sea al menos el 80% central de la imagen
		refinedBox := BBox{
			MinX: int(math.Min(0.1*w, left)),
			MinY: int(math.Min(0.1*h, upper)),
			MaxX: int(math.Max(0.9*w, right)),
			MaxY: int(math.Max(0.9*h, lower)),
		}

		// 4. Llamar a maybeCrop con el BBox refinado
		MaybeCrop(payload.Image, refinedBox, minimum, preserveMargin)
	}
}
