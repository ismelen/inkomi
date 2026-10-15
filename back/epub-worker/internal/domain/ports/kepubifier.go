package ports

import "context"

type Kepubifier interface {
	Kepubify(ctx context.Context, inputPath, outputPath string) error
}
