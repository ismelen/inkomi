package manga

// Profile describes the target e-reader device dimensions and format.
// Duplicated here from convert to avoid an import cycle (convert imports manga for Chapter).
// Both types must stay in sync.
type Profile struct {
	Width, Height int
	IsKepub       bool
}
