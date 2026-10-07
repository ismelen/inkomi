package usecases

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type UploadDoneUC struct {
	sourceRepo ports.SourceRepository
	queue      ports.Queue
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
		subject += "step.kepub"
	} else {
		subject += "init.manga"
	}

	err = u.queue.Publish(ctx, subject, bytes)
	if err != nil {
		return err
	}

	return u.sourceRepo.UpdateStatus(ctx, sourceId, userId, models.SourceStatusQueued, nil)
}

func (u *UploadDoneUC) isKepub(filename string) bool {
	return strings.Contains(filename, ".kepub")
}

func (u *UploadDoneUC) isEpub(filename string) bool {
	return filepath.Ext(filename) == ".epub"
}
