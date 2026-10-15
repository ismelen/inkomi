package downloadbook_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/downloadbook"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validMD5 = "0123456789abcdef0123456789abcdef"

type published struct {
	subject string
	msg     jobs.Msg
}

type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func recordingQueue(t *testing.T, out *[]published) *mocks.MockQueue {
	t.Helper()
	return &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload any) error {
			*out = append(*out, published{subject: subject, msg: payload.(jobs.Msg)})
			return nil
		},
	}
}

func libraryWith(title, ext, content string) *mocks.MockLibrary {
	return &mocks.MockLibrary{
		DownloadFn: func(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
			return &models.BookDownload{Stream: io.NopCloser(strings.NewReader(content)), Title: title, Ext: ext}, nil
		},
	}
}

func TestDownloadBookUC_Execute_InvalidMD5_ReturnsPermanent(t *testing.T) {
	t.Parallel()

	// Arrange
	var events []published
	downloaded := false
	library := &mocks.MockLibrary{
		DownloadFn: func(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
			downloaded = true
			return nil, nil
		},
	}
	uc := downloadbook.NewDownloadBookUC(recordingQueue(t, &events), library, &mocks.MockCloudStorage{})

	// Act
	err := uc.Execute(context.Background(), jobs.Msg{Id: "s1", Filename: "book.epub"}, jobrun.Delivery{})

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidMD5)
	assert.True(t, jobrun.IsPermanent(err))
	assert.False(t, downloaded)
	assert.Empty(t, events)
}

func TestDownloadBookUC_Execute_DownloadFails_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		err       error
		permanent bool
	}{
		{"Not found", models.ErrBookNotFound, true},
		{"No mirror", models.ErrNoMirror, false},
		{"Network", errors.New("connection reset"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var events []published
			library := &mocks.MockLibrary{
				DownloadFn: func(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
					return nil, tc.err
				},
			}
			uc := downloadbook.NewDownloadBookUC(recordingQueue(t, &events), library, &mocks.MockCloudStorage{})

			// Act
			err := uc.Execute(context.Background(), jobs.Msg{Id: "s1", Filename: validMD5}, jobrun.Delivery{})

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.permanent, jobrun.IsPermanent(err))
			require.Len(t, events, 1)
			assert.Equal(t, string(jobs.SubjectStart), events[0].subject)
		})
	}
}

func TestDownloadBookUC_Execute_UploadFails_ReturnsTransientAndClosesStream(t *testing.T) {
	t.Parallel()

	// Arrange
	var events []published
	stream := &closeTracker{Reader: strings.NewReader("content")}
	library := &mocks.MockLibrary{
		DownloadFn: func(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
			return &models.BookDownload{Stream: stream, Title: "Dune", Ext: "epub"}, nil
		},
	}
	storage := &mocks.MockCloudStorage{
		UploadFn: func(ctx context.Context, id string, filename string, body io.Reader) error {
			return errors.New("r2 down")
		},
	}
	uc := downloadbook.NewDownloadBookUC(recordingQueue(t, &events), library, storage)

	// Act
	err := uc.Execute(context.Background(), jobs.Msg{Id: "s1", Filename: validMD5}, jobrun.Delivery{})

	// Assert
	assert.Error(t, err)
	assert.False(t, jobrun.IsPermanent(err))
	assert.True(t, stream.closed)
	assert.Len(t, events, 1)
}

func TestDownloadBookUC_Execute_Downloaded_UploadsAndPublishesNextStep(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		ext      string
		msg      jobs.Msg
		subject  jobs.Subject
		filename string
	}{
		{"Epub", "epub", jobs.Msg{}, jobs.SubjectDone, "Dune.epub"},
		{"Epub to kepubify", "epub", jobs.Msg{Kepubify: true}, jobs.SubjectKepub, "Dune.epub"},
		{"Epub in a joined folder", "epub", jobs.Msg{Kepubify: true, Parent: &jobs.Parent{Id: "p1", Total: 2}}, jobs.SubjectJoin, "Dune.epub"},
		{"Pdf", "pdf", jobs.Msg{Kepubify: true, Parent: &jobs.Parent{Id: "p1", Total: 2}}, jobs.SubjectDone, "Dune.pdf"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var events []published
			var uploadedId, uploadedName, uploadedBody string
			storage := &mocks.MockCloudStorage{
				UploadFn: func(ctx context.Context, id string, filename string, body io.Reader) error {
					bytes, err := io.ReadAll(body)
					uploadedId, uploadedName, uploadedBody = id, filename, string(bytes)
					return err
				},
			}
			uc := downloadbook.NewDownloadBookUC(recordingQueue(t, &events), libraryWith("Dune", tc.ext, "book content"), storage)

			msg := tc.msg
			msg.Id, msg.UserId, msg.Filename = "s1", 7, strings.ToUpper(validMD5)

			// Act
			err := uc.Execute(context.Background(), msg, jobrun.Delivery{})

			// Assert
			require.NoError(t, err)
			assert.Equal(t, "s1", uploadedId)
			assert.Equal(t, tc.filename, uploadedName)
			assert.Equal(t, "book content", uploadedBody)

			require.Len(t, events, 2)
			assert.Equal(t, string(jobs.SubjectStart), events[0].subject)
			assert.Equal(t, string(tc.subject), events[1].subject)
			next := events[1].msg
			assert.Equal(t, "s1", next.Id)
			assert.Equal(t, 7, next.UserId)
			assert.Equal(t, tc.filename, next.Filename)
			assert.Equal(t, int64(len("book content")), next.Size)
			assert.Equal(t, "Dune", next.Title)
			assert.Equal(t, tc.msg.Parent, next.Parent)
		})
	}
}

func TestDownloadBookUC_Execute_MsgWithTitle_KeepsTitle(t *testing.T) {
	t.Parallel()

	// Arrange
	var events []published
	uc := downloadbook.NewDownloadBookUC(recordingQueue(t, &events), libraryWith("Dune", "epub", "x"), &mocks.MockCloudStorage{})

	// Act
	err := uc.Execute(context.Background(), jobs.Msg{Id: "s1", Filename: validMD5, Title: "My Dune"}, jobrun.Delivery{})

	// Assert
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "My Dune", events[1].msg.Title)
	assert.Equal(t, "Dune.epub", events[1].msg.Filename)
}
