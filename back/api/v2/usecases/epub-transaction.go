package usecases

import (
	"archive/zip"
	"context"
	"fmt"
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/cloud"
	filesHelper "ismelen/ermc/v2/infra/files-helper"
	"ismelen/ermc/v2/infra/state"
	"ismelen/ermc/v2/ports"
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
	stateManager := state.GetManager()
	stateManager.StartTransaction(config.Id, dstPath, filesHelper.GetSize(src))

	profile, err := domain.NewProfile(config.Profile)
	if err != nil {
		e.handleError(stateManager, config, err)
		return
	}

	if profile.IsKepub {
		kSrc, err := ConvertToKepub(src, dstPath, config.Title)
		if err != nil {
			e.handleError(stateManager, config, err)
			return
		}
		os.RemoveAll(src)
		src = kSrc
	}

	stateManager.SetResultPath(config.Id, src)

	if config.Cloud {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(src)))
		gCloud, err := cloud.New(config.CloudToken, config.CloudFolder)

		if err != nil {
			e.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(src)))
			return
		}
		gCloud.Upload(src)
	} else {
		e.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s transaction ready", filepath.Base(src)))
	}

	stateManager.SetDone(config.Id)
}

func (e *EpubTransactionUC) handleError(
	stateManager *state.TransactionStateManager,
	config *domain.TransactionConfig,
	err error,
) {
	e.pushNotifier.Send(config.NotifyToken, "Error", err.Error())
	stateManager.SetError(config.Id, err)
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
