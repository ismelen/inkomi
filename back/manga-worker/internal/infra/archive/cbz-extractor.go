package archive

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/facette/natsort"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

var _ ports.IPageExtractor = (*CBZExtractor)(nil)

var imageExts = []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tif", ".tiff"}

type CBZExtractor struct{}

func NewCBZExtractor() *CBZExtractor { return &CBZExtractor{} }

// Extract writes the images of the zip as dstDir/0001.ext, 0002.ext... in natural order of
// their path inside the archive. Entry names are never used as paths, so they can't escape dstDir.
func (c *CBZExtractor) Extract(ctx context.Context, srcPath string, dstDir string) ([]string, error) {
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		return nil, fmt.Errorf("open cbz: %w", err)
	}
	defer reader.Close()

	entries := make([]*zip.File, 0, len(reader.File))
	for _, file := range reader.File {
		if isPage(file) {
			entries = append(entries, file)
		}
	}
	slices.SortStableFunc(entries, func(a, b *zip.File) int {
		switch {
		case natsort.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)):
			return -1
		case natsort.Compare(strings.ToLower(b.Name), strings.ToLower(a.Name)):
			return 1
		default:
			return 0
		}
	})

	paths := make([]string, 0, len(entries))
	for i, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		dst := filepath.Join(dstDir, fmt.Sprintf("%04d%s", i+1, strings.ToLower(path.Ext(entry.Name))))
		if err := extractEntry(entry, dst); err != nil {
			return nil, fmt.Errorf("extract %s: %w", entry.Name, err)
		}
		paths = append(paths, dst)
	}
	return paths, nil
}

func isPage(file *zip.File) bool {
	if file.FileInfo().IsDir() {
		return false
	}
	name := strings.ReplaceAll(file.Name, "\\", "/")
	if strings.HasPrefix(name, "__MACOSX/") || strings.Contains(name, "/__MACOSX/") || strings.HasPrefix(path.Base(name), ".") {
		return false
	}
	return slices.Contains(imageExts, strings.ToLower(path.Ext(name)))
}

func extractEntry(entry *zip.File, dst string) error {
	src, err := entry.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, src); err != nil {
		// the copy error is the relevant one
		_ = out.Close()
		return err
	}
	return out.Close()
}
