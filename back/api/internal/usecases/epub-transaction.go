package usecases

import (
	"archive/zip"
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/convert"
	"os"
	"path/filepath"

	"github.com/pgaskin/kepubify/v4/kepub"
)

type EpubTransactionUC struct {
	pushNotifier convert.PushNotifier
	tranStore    convert.TransactionStore
	cloudStorage convert.CloudStorage
}

func NewEpubTransactionUC(
	pushNotifier convert.PushNotifier,
	tranStore convert.TransactionStore,
	cloudStorage convert.CloudStorage,
) *EpubTransactionUC {
	return &EpubTransactionUC{
		pushNotifier: pushNotifier,
		tranStore:    tranStore,
		cloudStorage: cloudStorage,
	}
}

func (e *EpubTransactionUC) Execute(src string, config *convert.TransactionConfig, dstPath string) {
	tran := e.tranStore.StartTransaction(config.Id, dstPath, 1)

	profile, err := convert.NewProfile(config.Profile)
	if err != nil {
		e.handleError(config, err)
		tran.SetError(err)
		return
	}

	if profile.IsKepub {
		kSrc, err := ConvertToKepub(src, dstPath, config.Title)
		if err != nil {
			e.handleError(config, err)
			tran.SetError(err)
			return
		}
		os.RemoveAll(src)
		src = kSrc
	}

	tran.SetResultPath(src)

	if config.Cloud && e.cloudStorage != nil {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(src)))
		if err := e.cloudStorage.Upload(src); err != nil {
			e.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(src)))
			tran.SetError(err)
			return
		}
	} else {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s transaction ready", filepath.Base(src)))
	}

	tran.SetDone()
}

func (e *EpubTransactionUC) handleError(config *convert.TransactionConfig, err error) {
	e.pushNotifier.Send(config.NotifyToken, "Error", err.Error())
}

func ConvertToKepub(src, outBase, filename string) (string, error) {
	kPath := filepath.Join(outBase, filename+".kepub.epub")
	out, err := os.Create(kPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	in, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	converter := kepub.NewConverter()
	ctx := context.Background()

	return kPath, converter.Convert(ctx, out, in)
}
