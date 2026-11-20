package mocks

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// MockCloudStorage defaults UploadFile and DownloadFile to Upload and Download over the local
// file, so a test can stub only the stream functions.
type MockCloudStorage struct {
	UploadFn       func(ctx context.Context, id string, filename string, body io.Reader) error
	DownloadFn     func(ctx context.Context, id string, filename string) (io.ReadCloser, error)
	UploadFileFn   func(ctx context.Context, id string, filename string, path string) (int64, error)
	DownloadFileFn func(ctx context.Context, id string, filename string, path string) error
	SizeFn         func(ctx context.Context, id string, filename string) (int64, error)
	DeleteFn       func(ctx context.Context, id string, filename string) error
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

func (m *MockCloudStorage) UploadFile(ctx context.Context, id string, filename string, path string) (int64, error) {
	if m.UploadFileFn != nil {
		return m.UploadFileFn(ctx, id, filename, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	if err := m.Upload(ctx, id, filename, file); err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (m *MockCloudStorage) DownloadFile(ctx context.Context, id string, filename string, path string) error {
	if m.DownloadFileFn != nil {
		return m.DownloadFileFn(ctx, id, filename, path)
	}

	reader, err := m.Download(ctx, id, filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, reader); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
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
