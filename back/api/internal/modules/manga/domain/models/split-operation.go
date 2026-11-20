package models

// SplitOperation describes how a page spread was split or rotated for the e-reader.
type SplitOperation int

const (
	SplitNone    SplitOperation = iota // single page
	SplitRotated                       // spread rotated to fit the screen
	SplitToRight                       // right half of a spread
	SplitToLeft                        // left half of a spread
)
