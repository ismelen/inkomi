package ports

import "context"

type IPageExtractor interface {
	// Extract writes the image pages of the archive at srcPath into dstDir and returns their
	// paths in reading order.
	Extract(ctx context.Context, srcPath string, dstDir string) ([]string, error)
}
