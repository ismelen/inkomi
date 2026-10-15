package kepubify

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/cloudfile"
)

type KepubifyUC struct {
	queue        ports.Queue
	cloudStorage ports.CloudStorage
	kepubifier   ports.Kepubifier
}

func NewKepubifyUC(queue ports.Queue, cloudStorage ports.CloudStorage, kepubifier ports.Kepubifier) *KepubifyUC {
	return &KepubifyUC{
		queue:        queue,
		cloudStorage: cloudStorage,
		kepubifier:   kepubifier,
	}
}

// Execute replaces the epub of msg in cloud storage with its kepub and publishes it as done.
func (k *KepubifyUC) Execute(ctx context.Context, msg jobs.Msg, _ jobrun.Delivery) error {
	if err := models.CheckEpubFilename(msg.Filename); err != nil {
		return jobrun.Permanent(fmt.Errorf("%s: %w", msg.Filename, err))
	}

	dir, err := os.MkdirTemp("", "kepubify-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		// best effort: a leftover temp dir doesn't affect the result
		_ = os.RemoveAll(dir)
	}()

	original := msg.Filename
	kepubName := models.KepubFilename(original)
	inputPath := filepath.Join(dir, "input.epub")
	if err := cloudfile.Download(ctx, k.cloudStorage, msg.Id, original, inputPath); err != nil {
		if k.alreadyConverted(ctx, msg.Id, kepubName) {
			return nil
		}
		return err
	}

	outputPath := filepath.Join(dir, "output"+models.KepubExt)
	if err := k.kepubifier.Kepubify(ctx, inputPath, outputPath); err != nil {
		return jobrun.Permanent(err)
	}

	size, err := cloudfile.Upload(ctx, k.cloudStorage, msg.Id, kepubName, outputPath)
	if err != nil {
		return err
	}

	msg.Filename, msg.Size = kepubName, size
	if err := k.queue.Publish(ctx, string(jobs.SubjectDone), msg); err != nil {
		return err
	}

	// The kepub is already published: retrying because the original couldn't be deleted
	// would redo the whole conversion, so the failure is only logged.
	if err := k.cloudStorage.Delete(ctx, msg.Id, original); err != nil {
		log.Printf("kepubify %s: delete %s: %v", msg.Id, original, err)
	}
	return nil
}

// alreadyConverted reports whether a redelivered job was finished by a previous attempt
// that deleted the original after uploading the kepub.
func (k *KepubifyUC) alreadyConverted(ctx context.Context, id, kepubName string) bool {
	_, err := k.cloudStorage.Size(ctx, id, kepubName)
	return err == nil
}
