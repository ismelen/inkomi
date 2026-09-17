package usecases

import (
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/strutil"
	"ismelen/inkomi/internal/shared/uid"
	"os"
	"path/filepath"
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

func (m MD5UC) Process(file *convert.TransactionFile, tran *convert.Transaction, transPath string) *convert.TransactionResultFile {
	result, err := m.downloadBookUC.Execute(file.SrcPath, 3)
	if err != nil {
		file.SetError(err)
		return nil
	}
	defer result.Stream.Close()

	safeTitle := strutil.SanitizeFilename(file.Name)
	result.Filename = safeTitle + "." + result.Ext

	dstPath := filepath.Join(transPath, tran.Id, file.Id, result.Filename)
	src, err := fs.CopyFromStream(result.Stream, dstPath)
	if err != nil {
		file.SetError(err)
		os.RemoveAll(src)
		return nil
	}

	return convert.NewTransactionResultFile(uid.GetRandomID(6), result.Filename, dstPath, 0, []*convert.TransactionFile{file})
}
