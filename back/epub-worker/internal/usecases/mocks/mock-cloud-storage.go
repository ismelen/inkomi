package mocks

import (
	"context"
	"io"
	"strings"
)

type MockCloudStorage struct {
	UploadFn   func(ctx context.Context, id string, filename string, body io.Reader) error
	DownloadFn func(ctx context.Context, id string, filename string) (io.ReadCloser, error)
	SizeFn     func(ctx context.Context, id string, filename string) (int64, error)
	DeleteFn   func(ctx context.Context, id string, filename string) error
}

func (m *MockCloudStorage) Upload(ctx context.Context, id string, filename string, body io.Reader) error {
	if m.UploadFn != nil {
		return m.UploadFn(ctx, id, filename, body)
	}
	return nil
}

func (m *MockCloudStorage) Download(ctx context.Context, id string, filename string) (io.ReadCloser, error) {
	if m.DownloadFn != nil {
		return m.DownloadFn(ctx, id, filename)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockCloudStorage) Size(ctx context.Context, id string, filename string) (int64, error) {
	if m.SizeFn != nil {
		return m.SizeFn(ctx, id, filename)
	}
	return 0, nil
}

func (m *MockCloudStorage) Delete(ctx context.Context, id string, filename string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, filename)
	}
	return nil
}
