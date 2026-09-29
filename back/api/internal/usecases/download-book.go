package usecases

import (
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/book"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type DownloadBookUC struct {
	provider book.BooksProvider
}

func NewDownloadBookUC(provider book.BooksProvider) *DownloadBookUC {
	return &DownloadBookUC{
		provider: provider,
	}
}

func (d *DownloadBookUC) Execute(ctx context.Context, md5 string, retries int) (*book.LibgenDownload, error) {
	ctx, span := otel.Tracer("inkomi-api").Start(ctx, "DownloadBookUC.Execute")
	defer span.End()

	if retries <= 0 {
		err := fmt.Errorf("download failed after %d retries, no working mirror", retries)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	mirror, ok := d.provider.GetMirror()
	if !ok {
		err := fmt.Errorf("no mirror available yet")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	resp, err := mirror.Download(md5)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		// d.refreshMirror()
		return d.Execute(ctx, md5, retries-1)
	}

	return resp, nil
}
