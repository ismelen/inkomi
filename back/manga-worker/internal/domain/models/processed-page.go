package models

// ProcessedPage is an encoded page ready for the e-reader. A source page may produce two
// when a spread is split.
type ProcessedPage struct {
	Data   []byte
	Width  int
	Height int
	Split  SplitOperation
	Format ImageFormat
}
