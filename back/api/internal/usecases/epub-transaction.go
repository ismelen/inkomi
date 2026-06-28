package usecases

import (
	"archive/zip"
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain"
	"ismelen/inkomi/internal/infra/cloud"
	"ismelen/inkomi/internal/infra/state"
	"ismelen/inkomi/internal/ports"
	"os"
	"path/filepath"

	"github.com/pgaskin/kepubify/v4/kepub"
)

type EpubTransactionUC struct {
	pushNotifier ports.PushNotifier
}

func NewEpubTransactionUC(pushNotifier ports.PushNotifier) *EpubTransactionUC {
	return &EpubTransactionUC{
		pushNotifier: pushNotifier,
	}
}

func (e *EpubTransactionUC) Execute(src string, config *domain.TransactionConfig, dstPath string) {
	transactionManager := state.GetManager()
	tran := transactionManager.StartTransaction(config.Id, dstPath, 1)

	profile, err := domain.NewProfile(config.Profile)
	if err != nil {
		e.handleError(tran, config, err)
		return
	}

	if profile.IsKepub {
		kSrc, err := ConvertToKepub(src, dstPath, config.Title)
		if err != nil {
			e.handleError(tran, config, err)
			return
		}
		os.RemoveAll(src)
		src = kSrc
	}

	tran.SetResultPath(src)

	if config.Cloud {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(src)))
		cloud, err := cloud.NewDropboxCloud(config.CloudToken, config.CloudFolder)

		if err != nil {
			e.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(src)))
			tran.SetError(err)
			return
		}
		if err := cloud.Upload(src); err != nil {
			e.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(src)))
			tran.SetError(err)
			return
		}
	} else {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s transaction ready", filepath.Base(src)))
	}

	tran.SetDone()
}

func (e *EpubTransactionUC) handleError(
	transaction *domain.Transaction,
	config *domain.TransactionConfig,
	err error,
) {
	e.pushNotifier.Send(config.NotifyToken, "Error", err.Error())
	transaction.SetError(err)
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
