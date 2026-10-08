package joinepubs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

type JoinEpubsUC struct {
	queue        ports.Queue
	cloudStorage ports.CloudStorage
	merger       ports.EpubMerger
}

func NewJoinEpubsUC(
	queue ports.Queue,
	cloudStorage ports.CloudStorage,
	merger ports.EpubMerger,
) *JoinEpubsUC {
	return &JoinEpubsUC{
		queue,
		cloudStorage,
		merger,
	}
}

func (j *JoinEpubsUC) Execute() error {
	return j.queue.Subscribe("job.step.join", func(ctx context.Context, subject string, payload []byte) error {
		var op models.JoinMsg
		if err := json.Unmarshal(payload, &op); err != nil {
			return j.queue.Publish(ctx, "job.step.error", map[string]any{
				"id":     op.Id,
				"error":  err.Error(),
				"userId": op.UserId,
			})
		}

		dir, paths, err := j.downloadAll(ctx, &op)
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		fileName := sanitizeFilename(op.Title) + ".epub"
		outputPath := filepath.Join(dir, fileName)
		if err := j.merger.Merge(paths, op.Title, "", outputPath); err != nil {
			return err
		}

		file, err := os.Open(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := j.cloudStorage.Upload(ctx, op.Id, fileName, file); err != nil {
			return err
		}

		if op.Kepubify {
			return j.queue.Publish(ctx, "job.step.kepub", map[string]any{
				"id":     op.Id,
				"userId": op.UserId,
			})
		}

		return j.queue.Publish(ctx, "job.step.done", map[string]any{
			"id":       op.Id,
			"userId":   op.UserId,
			"filename": fileName,
			"toCloud":  op.ToCloud,
		})

	})
}

func (j *JoinEpubsUC) downloadAll(ctx context.Context, op *models.JoinMsg) (string, []string, error) {
	tempDir, err := os.MkdirTemp("", "join-epubs-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	paths := []string{}

	for _, item := range op.Items {
		reader, err := j.cloudStorage.Download(ctx, item.Id, item.Filename)
		if err != nil {
			return tempDir, nil, fmt.Errorf("failed to download %s/%s: %w", item.Id, item.Filename, err)
		}

		destPath := filepath.Join(tempDir, item.Filename)
		destFile, err := os.Create(destPath)
		if err != nil {
			reader.Close()
			return tempDir, nil, fmt.Errorf("failed to create local file %s: %w", destPath, err)
		}

		_, err = io.Copy(destFile, reader)
		destFile.Close()
		reader.Close()

		if err != nil {
			return tempDir, nil, fmt.Errorf("failed to write to local file %s: %w", destPath, err)
		}

		paths = append(paths, destPath)
	}

	return tempDir, paths, nil
}

func sanitizeFilename(name string) string {
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]+`)
	sanitized := invalidChars.ReplaceAllString(name, "_")
	sanitized = strings.TrimSpace(sanitized)
	if sanitized == "" {
		return "merged"
	}
	return sanitized
}
