package kepub

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pgaskin/kepubify/v4/kepub"
)

// KepubifyKepubifier converts epubs with github.com/pgaskin/kepubify.
type KepubifyKepubifier struct {
	converter *kepub.Converter
}

func NewKepubifyKepubifier() *KepubifyKepubifier {
	return &KepubifyKepubifier{converter: kepub.NewConverter()}
}

func (k *KepubifyKepubifier) Kepubify(ctx context.Context, inputPath, outputPath string) (err error) {
	zr, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", inputPath, err)
	}
	defer zr.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outputPath, err)
	}
	defer func() {
		err = errors.Join(err, out.Close())
		if err != nil {
			// the partial output is useless; the conversion error is what matters
			_ = os.Remove(outputPath)
		}
	}()

	if err := k.converter.Convert(ctx, out, zr); err != nil {
		return fmt.Errorf("kepubify %s: %w", inputPath, err)
	}
	return nil
}
