package usecases

import (
	"context"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/uid"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type MD5UC struct {
	BaseTransactionUC
	downloadBookUC *DownloadBookUC
}

func NewMd5TransactionUC(
	pushNotifier convert.PushNotifier,
	cloud convert.CloudStorage,
	downloadBookUC *DownloadBookUC,
) *MD5UC {
	t := &MD5UC{
		BaseTransactionUC: BaseTransactionUC{
			pushNotifier: pushNotifier,
			cloud:        cloud,
		},
		downloadBookUC: downloadBookUC,
	}

	t.processor = t
	return t
}

func (m MD5UC) Process(ctx context.Context, file *convert.TransactionFile, tran *convert.Transaction, transPath string) *convert.TransactionResultFile {
	ctx, span := otel.Tracer("inkomi-api").Start(ctx, "MD5UC.Process")
	defer span.End()

	result, err := m.downloadBookUC.Execute(ctx, file.SrcPath, 3)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		file.SetError(err)
		return nil
	}
	defer result.Stream.Close()

	dstPath := filepath.Join(transPath, tran.Id, file.Id, result.Filename)
	src, err := fs.CopyFromStream(result.Stream, dstPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		file.SetError(err)
		os.RemoveAll(src)
		return nil
	}

	return convert.NewTransactionResultFile(uid.GetRandomID(6), result.Filename, dstPath, 0, []*convert.TransactionFile{file})
}
