package manga

// SplitOperation describes how a page spread is split/rotated for e-reader display.
type SplitOperation = int

const (
	SplitNone    SplitOperation = iota // Single page, no split
	SplitRotated                       // Rotated 270°
	SplitToRight                       // Right half of a spread
	SplitToLeft                        // Left half of a spread
)
