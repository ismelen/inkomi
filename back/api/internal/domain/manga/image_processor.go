package manga

import "context"

// ProcessOptions engloba toda la configuración para el procesado de una página.
type ProcessOptions struct {
	// RotateDoublePage indica si se debe rotar 90 grados una página doble en lugar de cortarla por la mitad.
	RotateDoublePage bool
	Crop             CropOptions
	Resize           ResizeOptions
	Color            ColorOptions
	Dither           DitherOptions
	Sharpen          SharpenOptions
	Output           OutputOptions
}

// CropOptions define los parámetros para el recorte (auto/manual) de márgenes blancos o bordes indeseados.
type CropOptions struct {
	// Enabled activa o desactiva el paso de recorte por completo.
	Enabled   bool
	// Algorithm especifica la estrategia de recorte: "auto" para detectar bordes por contenido o "manual" para usar los márgenes definidos.
	Algorithm string
	// Tolerance determina la sensibilidad en modo "auto" (por ejemplo, número de píxeles oscuros a tolerar antes de considerar que es el borde real).
	Tolerance int
	// Margins define los márgenes estáticos a recortar (en píxeles) cuando el algoritmo es "manual".
	Margins   struct {
		Top, Right, Bottom, Left int
	}
}

// ResizeOptions define los parámetros para redimensionar o ajustar la imagen a la resolución de destino.
type ResizeOptions struct {
	// TargetWidth es el ancho objetivo (ej: 1448). Si es 0, no fuerza el ancho.
	TargetWidth  int
	// TargetHeight es el alto objetivo (ej: 1072). Si es 0, no fuerza el alto.
	TargetHeight int
	// Interpolator define el algoritmo matemático para crear píxeles. Usa int para mapear a bimg.Interpolator (ej: bimg.Lanczos3) sin importar bimg.
	Interpolator int
	// Enlarge permite estirar la imagen si es más pequeña que las dimensiones de destino. Si es false, no la estirará.
	Enlarge      bool
	// Embed rellena el espacio sobrante con color de fondo (blanco/negro) si el aspect ratio de la imagen no coincide con el destino.
	Embed        bool
}

// ColorOptions define los ajustes de color y niveles, optimizando la imagen para el contraste de tinta electrónica.
type ColorOptions struct {
	// Enabled activa o desactiva las correcciones de color y niveles (ahorra mucho procesador si está en false).
	Enabled    bool
	// Gamma ajusta la curva de grises intermedios para aclarar u oscurecer sin perder blancos y negros absolutos.
	Gamma      float64
	// BlackPoint define el umbral mínimo (0-65535); cualquier color más oscuro que esto se convierte instantáneamente a negro puro.
	BlackPoint uint16
	// WhitePoint define el umbral máximo; cualquier color más claro que esto se quema a blanco puro (elimina ruido de fondo/papel viejo).
	WhitePoint uint16
	// Brightness ajusta el brillo general de la imagen.
	Brightness float64
	// Contrast ajusta la diferencia de intensidad general.
	Contrast   float64
	// Grayscale fuerza la conversión de toda la imagen a escala de grises.
	Grayscale  bool
}

// DitherOptions define los parámetros para la difusión de error (dithering) para pantallas con pocos bits de color.
type DitherOptions struct {
	// Enabled activa la difusión de puntos de tinta para simular niveles de gris.
	Enabled   bool
	// Algorithm es el nombre del algoritmo a aplicar (ej: "atkinson").
	Algorithm string
}

// SharpenOptions define los parámetros de la Máscara de Enfoque (Unsharp Mask) para hacer texto y bordes más nítidos.
type SharpenOptions struct {
	// Enabled activa o desactiva el enfoque.
	Enabled bool
	// Radius controla el tamaño del área o grosor del borde sobre el que se calcula el contraste.
	Radius  int
	// X1 define el umbral base o aplanamiento para el contraste local.
	X1      float64
	// Y2 define el nivel de oscurecimiento (darkening) aplicado a la zona interior del borde.
	Y2      float64
	// Y3 define el nivel de aclaramiento (lightening) aplicado a la zona exterior del halo del borde.
	Y3      float64
	// M1 pendiente base para sombras.
	M1      float64
	// M2 límite multiplicativo general del efecto sobre la nitidez de paso alto.
	M2      float64
}

// OutputOptions define los parámetros de compresión y codificación para el archivo binario resultante.
type OutputOptions struct {
	// Format define el tipo de archivo saliente. Usa int para mapear a bimg.ImageType (ej. PNG, JPEG, WEBP).
	Format  int
	// Palette (sólo PNG) fuerza el renderizado a color indexado 8-bits en vez de 24-bits, lo cual reduce drásticamente el peso del archivo.
	Palette bool
	// Quality ajusta la compresión (0-100) en formatos con pérdida (JPEG o WEBP).
	Quality int
}

// ImageProcessor defines the contract for processing manga pages
// for E-Ink displays.
type ImageProcessor interface {
	// ProcessPage processes a raw image byte slice, returning one or more
	// processed image byte slices (multiple in case of double page splitting).
	ProcessPage(ctx context.Context, imgData []byte, opts ProcessOptions) ([][]byte, error)
}
