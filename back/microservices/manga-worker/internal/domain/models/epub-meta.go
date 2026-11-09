package models

// EpubMeta describes a fixed-layout manga epub. Width and Height are the target screen size.
type EpubMeta struct {
	Title  string
	RTL    bool
	Width  int
	Height int
}

// EpubPage is a processed page stored on disk at Path.
type EpubPage struct {
	Path   string
	Width  int
	Height int
	Split  SplitOperation
	Format ImageFormat
}
