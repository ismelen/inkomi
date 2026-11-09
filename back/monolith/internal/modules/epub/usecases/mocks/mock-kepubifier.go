package mocks

import (
	"context"
	"os"
)

type MockKepubifier struct {
	KepubifyFn func(ctx context.Context, inputPath, outputPath string) error
}

// Kepubify writes an empty output file by default so the caller can upload it.
func (m *MockKepubifier) Kepubify(ctx context.Context, inputPath, outputPath string) error {
	if m.KepubifyFn != nil {
		return m.KepubifyFn(ctx, inputPath, outputPath)
	}
	return os.WriteFile(outputPath, []byte("kepub"), 0o600)
}
