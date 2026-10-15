package models

// EpubBook is one source epub of a merge; Title labels it in the merged table of contents.
type EpubBook struct {
	Path  string
	Title string
}

// EpubMeta holds the metadata of a merged epub. RTL also reverses the order of its books.
type EpubMeta struct {
	Title  string
	Author string
	RTL    bool
}
