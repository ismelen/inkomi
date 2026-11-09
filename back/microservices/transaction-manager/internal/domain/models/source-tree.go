package models

// SourceTree is a root source with its items and join results.
type SourceTree struct {
	Source
	Children []Source
}
