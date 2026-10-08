package ports

type EpubMerger interface {
	Merge(paths []string, title, author, outputPath string) error
}

