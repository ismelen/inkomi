package kepubify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/cloudfile"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/jobfail"
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

func (k *KepubifyUC) Execute() error {
	return k.queue.Subscribe(string(jobs.SubjectKepub), func(ctx context.Context, delivery ports.Delivery) error {
		return jobfail.Report(ctx, k.queue, delivery, k.handle(ctx, delivery.Payload))
	})
}

func (k *KepubifyUC) handle(ctx context.Context, payload []byte) error {
	var msg jobs.Msg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return jobfail.Permanent(fmt.Errorf("parse kepub job: %w", err))
	}
	if err := models.CheckEpubFilename(msg.Filename); err != nil {
		return jobfail.Permanent(fmt.Errorf("%s: %w", msg.Filename, err))
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
		return jobfail.Permanent(err)
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
