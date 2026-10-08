package uploaddone

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type UploadDoneUC struct {
	sourceRepo   ports.SourceRepository
	cloudStorage ports.CloudStorage
	queue        ports.Queue
}

func NewUploadDoneUC(
	sourceRepo ports.SourceRepository,
	cloudStorage ports.CloudStorage,
	queue ports.Queue,
) *UploadDoneUC {
	return &UploadDoneUC{
		sourceRepo,
		cloudStorage,
		queue,
	}
}

func (u *UploadDoneUC) Execute(ctx context.Context, userId int, sourceId string) error {
	src, err := u.sourceRepo.GetByIdAndUserIdCompact(ctx, sourceId, userId)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	subject := "job."
	if src.Type == models.SourceTypeLibrary {
		subject += "init.library"
	} else if u.isKepub(src.Filename) {
		subject += "step.done"
	} else if u.isEpub(src.Filename) {
		if src.Kepubify {
			subject += "step.kepub"
		} else {
			subject += "step.done"
		}
	} else {
		subject += "init.manga"
	}

	err = u.queue.Publish(ctx, subject, bytes)
	if err != nil {
		return err
	}

	if ok := u.cloudStorage.Check(src.Id, filepath.Ext(src.Filename)); !ok {
		return fmt.Errorf("source doesn't exists")
	}

	return u.sourceRepo.UpdateStatus(ctx, sourceId, userId, models.SourceStatusQueued, nil)
}

func (u *UploadDoneUC) isKepub(filename string) bool {
	return strings.Contains(filename, ".kepub")
}

func (u *UploadDoneUC) isEpub(filename string) bool {
	return filepath.Ext(filename) == ".epub"
}
