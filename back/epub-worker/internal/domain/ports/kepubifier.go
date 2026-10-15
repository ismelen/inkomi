package ports

import "context"

type IKepubifier interface {
	Kepubify(ctx context.Context, inputPath, outputPath string) error
}
