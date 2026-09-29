package usecases

import (
	"context"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/shared/uid"

	"go.opentelemetry.io/otel"
)

type EpubTransactionUC struct {
	BaseTransactionUC
	tranStore convert.TransactionStoreI
}

func NewEpubTransactionUC(
	pushNotifier convert.PushNotifier,
	cloud convert.CloudStorage,
) *EpubTransactionUC {
	t := &EpubTransactionUC{
		BaseTransactionUC: BaseTransactionUC{
			pushNotifier: pushNotifier,
			cloud:        cloud,
		},
	}

	t.processor = t
	return t
}

func (e EpubTransactionUC) Process(ctx context.Context, file *convert.TransactionFile, tran *convert.Transaction, transPath string) *convert.TransactionResultFile {
	ctx, span := otel.Tracer("inkomi-api").Start(ctx, "EpubTransactionUC.Process")
	defer span.End()

	return convert.NewTransactionResultFile(uid.GetRandomID(6), file.Name, file.SrcPath, file.Size, []*convert.TransactionFile{file})
}
