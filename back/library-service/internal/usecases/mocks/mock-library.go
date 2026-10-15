package mocks

import (
	"context"
	"io"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
)

type MockLibrary struct {
	SearchFn   func(ctx context.Context, query string) ([]models.Book, error)
	DownloadFn func(ctx context.Context, md5 models.MD5) (*models.BookDownload, error)
}

func (m *MockLibrary) Search(ctx context.Context, query string) ([]models.Book, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, query)
	}
	return nil, nil
}

func (m *MockLibrary) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
	if m.DownloadFn != nil {
		return m.DownloadFn(ctx, md5)
	}
	return &models.BookDownload{
		Stream: io.NopCloser(strings.NewReader("book")),
		Title:  "Book",
		Ext:    "epub",
	}, nil
}
