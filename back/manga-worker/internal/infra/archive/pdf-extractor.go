package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

var _ ports.IPageExtractor = (*PDFExtractor)(nil)

type PDFExtractor struct{}

func NewPDFExtractor() *PDFExtractor { return &PDFExtractor{} }

// Extract writes the images embedded in the pdf as dstDir/0001.ext, 0002.ext... in page order.
// Images are written as pdfcpu renders them (jpg, png, tif...), without re-encoding.
func (p *PDFExtractor) Extract(ctx context.Context, srcPath string, dstDir string) ([]string, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	defer file.Close()

	pages, err := api.ExtractImagesRaw(file, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("extract pdf images: %w", err)
	}

	var paths []string
	for _, images := range pages {
		objNrs := make([]int, 0, len(images))
		for objNr, img := range images {
			if !img.Thumb {
				objNrs = append(objNrs, objNr)
			}
		}
		slices.Sort(objNrs)

		for _, objNr := range objNrs {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			img := images[objNr]
			dst := filepath.Join(dstDir, fmt.Sprintf("%04d.%s", len(paths)+1, strings.ToLower(img.FileType)))
			if err := writeImage(img, dst); err != nil {
				return nil, fmt.Errorf("write pdf image %d: %w", objNr, err)
			}
			paths = append(paths, dst)
		}
	}
	return paths, nil
}

func writeImage(img model.Image, dst string) error {
	if closer, ok := img.Reader.(io.Closer); ok {
		defer closer.Close()
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, img); err != nil {
		// the copy error is the relevant one
		_ = out.Close()
		return err
	}
	return out.Close()
}
