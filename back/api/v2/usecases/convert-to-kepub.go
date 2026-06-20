package usecases

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"

	"github.com/pgaskin/kepubify/v4/kepub"
)

func ConvertToKepub(src, outBase, filename string) (string, error) {
	kPath := filepath.Join(outBase, filename+".kepub.epub")
	out, err := os.Create(kPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	in, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	converter := kepub.NewConverter()
	ctx := context.Background()

	return kPath, converter.Convert(ctx, out, in)
}
