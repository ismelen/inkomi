package senddone

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/ports"
)

type SendDoneUC struct {
	queue        queue.IQueue
	cloudStorage cloud.ICloudStorage
	userSettings ports.IUserSettings
	cloudDrive   ports.ICloudDrive
	notifier     ports.IPushNotifier
}

func NewSendDoneUC(
	queue queue.IQueue,
	cloudStorage cloud.ICloudStorage,
	userSettings ports.IUserSettings,
	cloudDrive ports.ICloudDrive,
	notifier ports.IPushNotifier,
) *SendDoneUC {
	return &SendDoneUC{
		queue:        queue,
		cloudStorage: cloudStorage,
		userSettings: userSettings,
		cloudDrive:   cloudDrive,
		notifier:     notifier,
	}
}

// Execute uploads the file of a done job to the user's Dropbox when requested, notifies the
// user and publishes the job as sent.
func (s *SendDoneUC) Execute(ctx context.Context, msg jobs.Msg, _ jobrun.Delivery) error {
	if msg.ToCloud {
		if err := s.sendToCloud(ctx, msg); err != nil {
			return err
		}
	}

	s.notify(ctx, msg)

	return s.queue.Publish(ctx, string(jobs.SubjectSent), msg)
}

func (s *SendDoneUC) sendToCloud(ctx context.Context, msg jobs.Msg) error {
	account, err := s.userSettings.GetDropboxAccount(ctx, msg.UserId)
	if errors.Is(err, models.ErrDropboxNotLinked) {
		return jobrun.Permanent(err)
	}
	if err != nil {
		return fmt.Errorf("get dropbox account: %w", err)
	}

	dir, err := os.MkdirTemp("", "send-done-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		// best effort: a leftover temp dir doesn't affect the result
		_ = os.RemoveAll(dir)
	}()

	path := filepath.Join(dir, "file")
	if err := s.cloudStorage.DownloadFile(ctx, msg.Id, msg.Filename, path); err != nil {
		return fmt.Errorf("download %s: %w", msg.Filename, err)
	}

	err = s.cloudDrive.UploadFile(ctx, account, path, msg.Filename)
	if errors.Is(err, models.ErrUploadRejected) {
		return jobrun.Permanent(err)
	}
	if err != nil {
		return fmt.Errorf("upload %s: %w", msg.Filename, err)
	}
	return nil
}

// notify is best effort: a missing device or a failed push must not block the job.
func (s *SendDoneUC) notify(ctx context.Context, msg jobs.Msg) {
	token, err := s.userSettings.GetFCMToken(ctx, msg.UserId)
	if errors.Is(err, models.ErrFCMTokenNotFound) {
		return
	}
	if err != nil {
		log.Printf("send-done %s: get fcm token: %v", msg.Id, err)
		return
	}

	if err := s.notifier.Send(ctx, token, models.NewSentMessage(msg)); err != nil {
		log.Printf("send-done %s: push: %v", msg.Id, err)
	}
}
