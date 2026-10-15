package mocks

import "os"

type MockEpubMerger struct {
	MergeFn func(paths []string, title, author, outputPath string) error
}

// Merge writes an empty output file by default so the caller can upload it.
func (m *MockEpubMerger) Merge(paths []string, title, author, outputPath string) error {
	if m.MergeFn != nil {
		return m.MergeFn(paths, title, author, outputPath)
	}
	return os.WriteFile(outputPath, nil, 0o600)
}
