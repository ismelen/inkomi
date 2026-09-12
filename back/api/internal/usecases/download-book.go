package usecases

import (
	"fmt"
	"ismelen/inkomi/internal/domain/book"
)

type DownloadBookUC struct {
	provider book.BooksProvider
}

func NewDownloadBookUC(provider book.BooksProvider) *DownloadBookUC {
	return &DownloadBookUC{
		provider: provider,
	}
}

func (d *DownloadBookUC) Execute(md5 string, retries int) (*book.LibgenDownload, error) {
	if retries <= 0 {
		return nil, fmt.Errorf("download failed, no working mirror")
	}

	mirror, ok := d.provider.GetMirror()
	if !ok {
		return nil, fmt.Errorf("no mirror available yet")
	}

	resp, err := mirror.Download(md5)
	if err != nil {
		if updated := d.provider.Refresh(); !updated {
			return nil, fmt.Errorf("no working mirrors")
		}
		return d.Execute(md5, retries-1)
	}

	return resp, nil
}
