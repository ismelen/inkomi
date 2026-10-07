package manga

import "context"

// ProcessOptions engloba toda la configuración para el procesado de una página.
type ProcessOptions struct {
	// RotateDoublePage indica si se debe rotar 90 grados una página doble en lugar de cortarla por la mitad.
	RotateDoublePage bool
	// TargetWidth es el ancho objetivo (ej: 1448). Si es 0, usa 1448 por defecto.
	TargetWidth int
	// TargetHeight es el alto objetivo (ej: 1072). Si es 0, usa 1072 por defecto.
	TargetHeight int

	Crop    CropOptions
	Resize  ResizeOptions
	Color   ColorOptions
	Dither  DitherOptions
	Sharpen SharpenOptions
	Output  OutputOptions
}

// CropOptions define los parámetros para el recorte automático de márgenes blancos.
type CropOptions struct {
	Enabled   bool
	Tolerance float64
}

// ResizeOptions define parámetros adicionales para escalar la imagen.
type ResizeOptions struct {
	Enlarge bool
	Embed   bool
}

// ColorOptions define los ajustes de color y niveles.
type ColorOptions struct {
	Enabled bool
	// Gamma: 1.0 es neutral. < 1.0 aclara los tonos medios (ej. 0.8), > 1.0 los oscurece (ej. 1.2).
	Gamma float64
	// BlackPoint: Rango 0.0 a 1.0. Valores típicos 0.05 - 0.15. Corta los grises oscuros a negro puro.
	BlackPoint float64
	// WhitePoint: Rango 0.0 a 1.0. Valores típicos 0.90 - 0.98. Quema los grises claros a blanco puro.
	WhitePoint float64
	// Brightness: Brillo general. Un rango típico puede ser -1.0 a 1.0.
	Brightness float64
	// Contrast: Diferencia de intensidad. Rango típico 0.0 a 2.0 (1.0 neutral).
	Contrast float64
	// Grayscale: Fuerza a grises si es true.
	Grayscale bool
}

// DitherOptions activa la difusión de puntos de tinta para simular niveles de gris.
type DitherOptions struct {
	Enabled bool
}

// SharpenOptions define los parámetros de la Máscara de Enfoque (Unsharp Mask) para hacer texto y bordes más nítidos.
type SharpenOptions struct {
	Enabled bool
	Radius  int
	X1      float64
	Y2      float64
	Y3      float64
	M1      float64
	M2      float64
}

// OutputOptions define los parámetros de compresión y codificación para el archivo binario resultante.
type OutputOptions struct {
	Format  int  // Maps to bimg.ImageType
	Palette bool // Index to 8-bit
	Quality int
}

// NewDefaultProcessOptions devuelve la configuración base estándar del procesador de imágenes.
func NewDefaultProcessOptions() ProcessOptions {
	return ProcessOptions{
		RotateDoublePage: false,
		TargetWidth:      1448,
		TargetHeight:     1072,
		Crop: CropOptions{
			Enabled:   true,
			Tolerance: 0.02,
		},
		Resize: ResizeOptions{
			Enlarge: false,
			Embed:   false,
		},
		Color: ColorOptions{
			Enabled:    true,
			Gamma:      0.9,
			BlackPoint: 0.06, // equivale a ~15/255
			WhitePoint: 0.95, // equivale a ~242/255
			Brightness: 0.0,
			Contrast:   1.0,
			Grayscale:  true,
		},
		Dither: DitherOptions{
			Enabled: true,
		},
		Sharpen: SharpenOptions{
			Enabled: true,
			Radius:  1,
			X1:      1.5,
			Y2:      20,
			Y3:      50,
			M1:      0,
			M2:      3,
		},
		Output: OutputOptions{
			Format:  1, // 1 es PNG en bimg
			Palette: true,
			Quality: 100,
		},
	}
}

// ImageProcessor defines the contract for processing manga pages
// for E-Ink displays.
type ImageProcessor interface {
	// ProcessPage processes a raw image byte slice, returning one or more
	// processed image byte slices (multiple in case of double page splitting).
	ProcessPage(ctx context.Context, imgData []byte, opts ProcessOptions) ([][]byte, error)
}
