package downloadbook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/ports"
)

type DownloadBookUC struct {
	queue        queue.IQueue
	library      ports.ILibrary
	cloudStorage cloud.ICloudStorage
}

func NewDownloadBookUC(queue queue.IQueue, library ports.ILibrary, cloudStorage cloud.ICloudStorage) *DownloadBookUC {
	return &DownloadBookUC{
		queue:        queue,
		library:      library,
		cloudStorage: cloudStorage,
	}
}

// Execute downloads the book whose md5 is msg.Filename, stores it as msg.Id/<title>.<ext>
// and publishes it on the next step of the pipeline.
func (d *DownloadBookUC) Execute(ctx context.Context, msg jobs.Msg, _ jobrun.Delivery) error {
	md5, err := models.NewMD5(msg.Filename)
	if err != nil {
		return jobrun.Permanent(err)
	}

	if err := d.queue.Publish(ctx, string(jobs.SubjectStart), msg); err != nil {
		return err
	}

	book, err := d.library.Download(ctx, md5)
	if err != nil {
		if errors.Is(err, models.ErrBookNotFound) {
			return jobrun.Permanent(err)
		}
		return fmt.Errorf("download %s: %w", md5, err)
	}
	defer book.Stream.Close()

	filename, err := book.Filename()
	if err != nil {
		return jobrun.Permanent(fmt.Errorf("filename of %s: %w", md5, err))
	}

	dir, err := os.MkdirTemp("", "library-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		// best effort: a leftover temp dir doesn't affect the result
		_ = os.RemoveAll(dir)
	}()

	path := filepath.Join(dir, "book"+filepath.Ext(filename))
	if err := saveStream(book.Stream, path); err != nil {
		return err
	}

	size, err := d.cloudStorage.UploadFile(ctx, msg.Id, filename, path)
	if err != nil {
		return err
	}

	if msg.Title == "" {
		msg.Title = book.Title
	}
	msg.Filename, msg.Size = filename, size
	return d.queue.Publish(ctx, string(nextSubject(msg)), msg)
}
