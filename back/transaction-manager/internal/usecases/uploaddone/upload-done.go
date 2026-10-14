package uploaddone

import (
	"context"
	"encoding/json"
	"path/filepath"

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
		sourceRepo:   sourceRepo,
		cloudStorage: cloudStorage,
		queue:        queue,
	}
}

func (u *UploadDoneUC) Execute(ctx context.Context, userID int, sourceID string) error {
	src, err := u.sourceRepo.GetDetailByIDAndUserID(ctx, sourceID, userID)
	if err != nil {
		return err
	}

	if !src.IsLibrary() && !u.cloudStorage.Check(src.ID, filepath.Ext(src.Filename)) {
		return models.ErrSourceNotUploaded
	}

	subject, msg := u.buildJob(src)
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := u.queue.Publish(ctx, subject, bytes); err != nil {
		return err
	}

	return u.sourceRepo.UpdateStatus(ctx, sourceID, userID, models.SourceStatusQueued)
}

// buildJob picks the worker the source goes to and builds the message it expects:
// the base fields, plus manga settings for the manga worker and the parent folder
// when the source has to be joined.
func (u *UploadDoneUC) buildJob(src *models.SourceDetail) (string, models.JobMsg) {
	msg := models.JobMsg{
		ID:       src.ID,
		UserID:   src.UserID,
		Filename: src.Filename,
		Title:    src.Title,
		Kepubify: src.Kepubify,
		ToCloud:  src.ToCloud,
	}

	if src.Parent != nil && src.Parent.Join {
		msg.Parent = &models.JobParent{
			ID:    src.Parent.ID,
			Title: src.Parent.Title,
			RTL:   src.Parent.IsRTL,
		}
	}

	switch {
	case src.IsLibrary():
		return models.SubjectJobLibrary, msg
	case !src.IsEpub():
		msg.MangaJob = &models.MangaJob{RTL: src.IsRTL}
		if src.Config != nil {
			msg.MangaCfg = json.RawMessage(src.Config.Data)
		}
		return models.SubjectJobManga, msg
	case msg.Parent != nil:
		return models.SubjectJobWaitJoin, msg
	case src.Kepubify && !src.IsKepub():
		return models.SubjectJobKepub, msg
	default:
		return models.SubjectJobDone, msg
	}
}
