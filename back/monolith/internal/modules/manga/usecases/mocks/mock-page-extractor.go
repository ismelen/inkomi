package mocks

import "context"

type MockPageExtractor struct {
	ExtractFn func(ctx context.Context, srcPath string, dstDir string) ([]string, error)
}

func (m *MockPageExtractor) Extract(ctx context.Context, srcPath string, dstDir string) ([]string, error) {
	if m.ExtractFn != nil {
		return m.ExtractFn(ctx, srcPath, dstDir)
	}
	return nil, nil
}
