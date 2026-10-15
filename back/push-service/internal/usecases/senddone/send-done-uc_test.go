package senddone_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/push-service/internal/usecases/senddone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publishedEvent struct {
	subject string
	msg     jobs.Msg
}

type fixture struct {
	lastAttempt bool
	queue       *mocks.MockQueue
	cloud       *mocks.MockCloudStorage
	settings    *mocks.MockUserSettings
	drive       *mocks.MockCloudDrive
	notifier    *mocks.MockPushNotifier
	published   []publishedEvent
	downloaded  []string
	uploaded    []string
	pushed      []models.PushMessage
}

func newFixture() *fixture {
	f := &fixture{}
	f.queue = &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload any) error {
			f.published = append(f.published, publishedEvent{subject, payload.(jobs.Msg)})
			return nil
		},
	}
	f.cloud = &mocks.MockCloudStorage{
		DownloadFn: func(ctx context.Context, id, filename string) (io.ReadCloser, error) {
			f.downloaded = append(f.downloaded, id+"/"+filename)
			return io.NopCloser(strings.NewReader("epub")), nil
		},
	}
	f.settings = &mocks.MockUserSettings{
		GetDropboxAccountFn: func(ctx context.Context, userId int) (*models.DropboxAccount, error) {
			return models.NewDropboxAccount("dropbox-token", "/Inkomi")
		},
		GetFCMTokenFn: func(ctx context.Context, userId int) (string, error) {
			return "fcm-token", nil
		},
	}
	f.drive = &mocks.MockCloudDrive{
		UploadFileFn: func(ctx context.Context, account *models.DropboxAccount, path, filename string) error {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			f.uploaded = append(f.uploaded, account.Path(filename)+":"+string(content))
			return nil
		},
	}
	f.notifier = &mocks.MockPushNotifier{
		SendFn: func(ctx context.Context, token string, message models.PushMessage) error {
			f.pushed = append(f.pushed, message)
			return nil
		},
	}
	return f
}

// handle runs the use case the way main wires it: through jobrun.Handle.
func (f *fixture) handle(t *testing.T, msg jobs.Msg) error {
	t.Helper()

	payload, err := json.Marshal(msg)
	require.NoError(t, err)

	uc := senddone.NewSendDoneUC(f.queue, f.cloud, f.settings, f.drive, f.notifier)
	delivery := jobrun.Delivery{Subject: string(jobs.SubjectDone), Payload: payload, LastAttempt: f.lastAttempt}
	return jobrun.Handle(f.queue, uc.Execute)(context.Background(), delivery)
}

func job(toCloud bool) jobs.Msg {
	return jobs.Msg{Id: "job-1", UserId: 7, Filename: "dune.epub", Title: "Dune", ToCloud: toCloud}
}

func TestSendDoneUC_Execute_NotToCloud_NotifiesAndPublishesSent(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()

	// Act
	err := f.handle(t, job(false))

	// Assert
	require.NoError(t, err)
	assert.Empty(t, f.downloaded)
	assert.Empty(t, f.uploaded)
	require.Len(t, f.pushed, 1)
	assert.Equal(t, "Dune is ready", f.pushed[0].Body)
	require.Len(t, f.published, 1)
	assert.Equal(t, string(jobs.SubjectSent), f.published[0].subject)
	assert.Equal(t, job(false), f.published[0].msg)
}

func TestSendDoneUC_Execute_ToCloud_UploadsAndPublishesSent(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()

	// Act
	err := f.handle(t, job(true))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"job-1/dune.epub"}, f.downloaded)
	assert.Equal(t, []string{"/Inkomi/dune.epub:epub"}, f.uploaded)
	require.Len(t, f.pushed, 1)
	assert.Equal(t, "Dune was sent to Dropbox", f.pushed[0].Body)
	require.Len(t, f.published, 1)
	assert.Equal(t, string(jobs.SubjectSent), f.published[0].subject)
	assert.Equal(t, job(true), f.published[0].msg)
}

func TestSendDoneUC_Execute_PermanentCloudErrors_PublishesFailed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(f *fixture)
	}{
		{"Dropbox not linked", func(f *fixture) {
			f.settings.GetDropboxAccountFn = func(ctx context.Context, userId int) (*models.DropboxAccount, error) {
				return nil, models.ErrDropboxNotLinked
			}
		}},
		{"Upload rejected", func(f *fixture) {
			f.drive.UploadFileFn = func(ctx context.Context, account *models.DropboxAccount, path, filename string) error {
				return models.ErrUploadRejected
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture()
			tc.setup(f)

			// Act
			err := f.handle(t, job(true))

			// Assert
			require.NoError(t, err)
			assert.Empty(t, f.pushed)
			require.Len(t, f.published, 1)
			assert.Equal(t, string(jobs.SubjectFailed), f.published[0].subject)
			assert.NotEmpty(t, f.published[0].msg.ErrorMsg)
		})
	}
}

func TestSendDoneUC_Execute_TransientCloudErrors_ReturnsErrorForRetry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(f *fixture)
	}{
		{"Auth service down", func(f *fixture) {
			f.settings.GetDropboxAccountFn = func(ctx context.Context, userId int) (*models.DropboxAccount, error) {
				return nil, errors.New("connection refused")
			}
		}},
		{"Download fails", func(f *fixture) {
			f.cloud.DownloadFn = func(ctx context.Context, id, filename string) (io.ReadCloser, error) {
				return nil, errors.New("r2 unavailable")
			}
		}},
		{"Upload fails", func(f *fixture) {
			f.drive.UploadFileFn = func(ctx context.Context, account *models.DropboxAccount, path, filename string) error {
				return errors.New("dropbox 503")
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture()
			tc.setup(f)

			// Act
			err := f.handle(t, job(true))

			// Assert
			assert.Error(t, err)
			assert.Empty(t, f.pushed)
			assert.Empty(t, f.published)
		})
	}
}

func TestSendDoneUC_Execute_TransientErrorOnLastAttempt_PublishesFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()
	f.lastAttempt = true
	f.drive.UploadFileFn = func(ctx context.Context, account *models.DropboxAccount, path, filename string) error {
		return errors.New("dropbox 503")
	}

	// Act
	err := f.handle(t, job(true))

	// Assert
	require.NoError(t, err)
	require.Len(t, f.published, 1)
	assert.Equal(t, string(jobs.SubjectFailed), f.published[0].subject)
}

func TestSendDoneUC_Execute_PushFailures_StillPublishesSent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(f *fixture)
	}{
		{"No fcm token", func(f *fixture) {
			f.settings.GetFCMTokenFn = func(ctx context.Context, userId int) (string, error) {
				return "", models.ErrFCMTokenNotFound
			}
		}},
		{"Auth service down", func(f *fixture) {
			f.settings.GetFCMTokenFn = func(ctx context.Context, userId int) (string, error) {
				return "", errors.New("connection refused")
			}
		}},
		{"Push fails", func(f *fixture) {
			f.notifier.SendFn = func(ctx context.Context, token string, message models.PushMessage) error {
				return errors.New("unregistered token")
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture()
			tc.setup(f)

			// Act
			err := f.handle(t, job(true))

			// Assert
			require.NoError(t, err)
			require.Len(t, f.published, 1)
			assert.Equal(t, string(jobs.SubjectSent), f.published[0].subject)
		})
	}
}

func TestSendDoneUC_Execute_PublishFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()
	f.queue.PublishFn = func(ctx context.Context, subject string, payload any) error {
		return errors.New("nats down")
	}

	// Act
	err := f.handle(t, job(false))

	// Assert
	assert.Error(t, err)
}
