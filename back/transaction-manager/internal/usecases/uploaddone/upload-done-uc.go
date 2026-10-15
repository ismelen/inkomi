package uploaddone

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/ismelen/inkomi/back/common/jobs"
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

func (u *UploadDoneUC) Execute(ctx context.Context, userId int, sourceId string) error {
	src, err := u.sourceRepo.GetDetailByIdAndUserId(ctx, sourceId, userId)
	if err != nil {
		return err
	}

	if !src.IsLibrary() && !u.cloudStorage.Check(src.Id, filepath.Ext(src.Filename)) {
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

	return u.sourceRepo.UpdateStatus(ctx, sourceId, userId, models.SourceStatusQueued)
}

func (u *UploadDoneUC) buildJob(src *models.SourceDetail) (jobs.Subject, jobs.Msg) {
	msg := jobs.Msg{
		Id:       src.Id,
		UserId:   src.UserId,
		Filename: src.Filename,
		Title:    src.Title,
		Kepubify: src.Kepubify,
		ToCloud:  src.ToCloud,
	}
	if src.Size != nil {
		msg.Size = *src.Size
	}

	if src.Parent != nil && src.Parent.Join {
		msg.Parent = &jobs.Parent{
			Id:    src.Parent.Id,
			Title: src.Parent.Title,
			Rtl:   src.Parent.IsRTL,
			Total: src.Parent.ChildCount,
			Index: src.Position,
		}
	}

	switch {
	case src.IsLibrary():
		return jobs.SubjectLibrary, msg
	case !src.IsEpub():
		msg.Manga = &jobs.Manga{Rtl: src.IsRTL}
		if src.Config != nil {
			msg.MangaCfg = json.RawMessage(src.Config.Data)
		}
		return jobs.SubjectManga, msg
	case msg.Parent != nil:
		return jobs.SubjectJoin, msg
	case src.Kepubify && !src.IsKepub():
		return jobs.SubjectKepub, msg
	default:
		return jobs.SubjectDone, msg
	}
}
