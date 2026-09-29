package manga

import "context"

// ProcessOptions defines the parameters for processing a manga page.
type ProcessOptions struct {
	RotateDoublePage bool
	TargetWidth      int
	TargetHeight     int
}

// ImageProcessor defines the contract for processing manga pages
// for E-Ink displays.
type ImageProcessor interface {
	// ProcessPage processes a raw image byte slice, returning one or more
	// processed image byte slices (multiple in case of double page splitting).
	ProcessPage(ctx context.Context, imgData []byte, opts ProcessOptions) ([][]byte, error)
}
