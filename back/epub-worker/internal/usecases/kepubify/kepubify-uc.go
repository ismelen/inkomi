package kepubify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/cloudfile"
)

const kepubExt = ".kepub.epub"

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
	return k.queue.Subscribe(string(jobs.SubjectKepub), func(ctx context.Context, subject string, payload []byte) error {
		return k.handle(ctx, payload)
	})
}

func (k *KepubifyUC) handle(ctx context.Context, payload []byte) error {
	var msg jobs.Msg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return k.fail(ctx, msg, err)
	}
	if IsKepub(msg.Filename) {
		return k.queue.Publish(ctx, string(jobs.SubjectDone), msg)
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
	kepubName := KepubFilename(original)
	inputPath := filepath.Join(dir, "input.epub")
	if err := cloudfile.Download(ctx, k.cloudStorage, msg.Id, original, inputPath); err != nil {
		if k.alreadyConverted(ctx, msg.Id, kepubName) {
			return nil
		}
		return err
	}

	outputPath := filepath.Join(dir, "output"+kepubExt)
	if err := k.kepubifier.Kepubify(ctx, inputPath, outputPath); err != nil {
		return k.fail(ctx, msg, err)
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

// fail reports a job that can't be converted; returning the publish result acks the message so it isn't retried.
func (k *KepubifyUC) fail(ctx context.Context, msg jobs.Msg, err error) error {
	msg.ErrorMsg = err.Error()
	return k.queue.Publish(ctx, string(jobs.SubjectFailed), msg)
}

func IsKepub(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), kepubExt)
}

// KepubFilename turns "book.epub" into "book.kepub.epub".
func KepubFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + kepubExt
}
