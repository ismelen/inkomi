package usecases

import (
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/infra/cloud"
	"ismelen/inkomi/internal/infra/fs"
	"os"
	"path/filepath"
)

type RemoteTransactionUC struct {
	pushNotifier convert.PushNotifier
	tranStore    convert.TransactionStore
	libgenServ   book.LibgenService
}

func NewRemoteTransactionUC(
	pushNotifier convert.PushNotifier,
	tranStore convert.TransactionStore,
	libgenServ book.LibgenService,
) *RemoteTransactionUC {
	return &RemoteTransactionUC{
		pushNotifier: pushNotifier,
		tranStore:    tranStore,
		libgenServ:   libgenServ,
	}
}

func (e *RemoteTransactionUC) Execute(md5 string, config *convert.TransactionConfig, dstPath string) {
	tran := e.tranStore.StartTransaction(config.Id, dstPath, 1)

	result, err := e.libgenServ.Download(md5)
	if err != nil {
		result.Stream.Close()
		e.handleError(config, err)
		return
	}
	defer result.Stream.Close()

	src, err := fs.CopyFromStream(result.Stream, filepath.Join(dstPath, result.Filename))
	if err != nil {
		e.handleError(config, err)
		os.RemoveAll(src)
		return
	}

	profile, err := convert.NewProfile(config.Profile)
	if err != nil {
		e.handleError(config, err)
		tran.SetError(err)
		return
	}

	if profile.IsKepub {
		kSrc, err := ConvertToKepub(src, dstPath, result.Title)
		if err != nil {
			e.handleError(config, err)
			tran.SetError(err)
			return
		}
		os.RemoveAll(src)
		src = kSrc
	}

	tran.SetResultPath(src)

	if config.Cloud {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(src)))
		cld, _ := cloud.NewDropboxCloud(config.CloudToken, config.CloudFolder)

		if err := cld.Upload(src); err != nil {
			e.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(src)))
			tran.SetError(err)
			return
		}
	} else {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s transaction ready", filepath.Base(src)))
	}

	tran.SetDone()
}

func (e *RemoteTransactionUC) handleError(config *convert.TransactionConfig, err error) {
	e.pushNotifier.Send(config.NotifyToken, "Error", err.Error())
}
