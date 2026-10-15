package kepubify_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/kepubify"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/mocks"
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
	kepubifier  *mocks.MockKepubifier
	published   []publishedEvent
	uploaded    []string
	deleted     []string
}

func newFixture() *fixture {
	f := &fixture{kepubifier: &mocks.MockKepubifier{}}
	f.queue = &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload any) error {
			f.published = append(f.published, publishedEvent{subject, payload.(jobs.Msg)})
			return nil
		},
	}
	f.cloud = &mocks.MockCloudStorage{
		DownloadFn: func(ctx context.Context, id, filename string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("epub")), nil
		},
		UploadFn: func(ctx context.Context, id, filename string, body io.Reader) error {
			f.uploaded = append(f.uploaded, id+"/"+filename)
			return nil
		},
		DeleteFn: func(ctx context.Context, id, filename string) error {
			f.deleted = append(f.deleted, id+"/"+filename)
			return nil
		},
	}
	return f
}

func (f *fixture) handle(t *testing.T, payload []byte) error {
	t.Helper()

	var handler ports.EventHandler
	f.queue.SubscribeFn = func(subject string, h ports.EventHandler) error {
		assert.Equal(t, string(jobs.SubjectKepub), subject)
		handler = h
		return nil
	}

	require.NoError(t, kepubify.NewKepubifyUC(f.queue, f.cloud, f.kepubifier).Execute())
	require.NotNil(t, handler)
	return handler(context.Background(), ports.Delivery{Subject: string(jobs.SubjectKepub), Payload: payload, LastAttempt: f.lastAttempt})
}

func job(t *testing.T, filename string) []byte {
	t.Helper()

	bytes, err := json.Marshal(jobs.Msg{
		Id:       "result-1",
		UserId:   1,
		Filename: filename,
		Kepubify: true,
		Result:   &jobs.Result{ParentId: "folder-1", Index: 0, Total: 1},
	})
	require.NoError(t, err)
	return bytes
}

func TestKepubifyUC_Execute_Epub_ReplacesFileAndPublishesDone(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()

	// Act
	err := f.handle(t, job(t, "book.epub"))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"result-1/book.kepub.epub"}, f.uploaded)
	assert.Equal(t, []string{"result-1/book.epub"}, f.deleted)
	require.Len(t, f.published, 1)
	done := f.published[0]
	assert.Equal(t, string(jobs.SubjectDone), done.subject)
	assert.Equal(t, "book.kepub.epub", done.msg.Filename)
	assert.Equal(t, int64(len("kepub")), done.msg.Size)
	assert.Equal(t, &jobs.Result{ParentId: "folder-1", Index: 0, Total: 1}, done.msg.Result)
}

func TestKepubifyUC_Execute_InvalidJob_PublishesFailed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload func(t *testing.T) []byte
		setup   func(f *fixture)
	}{
		{"Invalid json", func(t *testing.T) []byte { return []byte("{") }, func(f *fixture) {}},
		{"Already kepub", func(t *testing.T) []byte { return job(t, "book.kepub.epub") }, func(f *fixture) {}},
		{"Not an epub", func(t *testing.T) []byte { return job(t, "book.pdf") }, func(f *fixture) {}},
		{"Conversion fails", func(t *testing.T) []byte { return job(t, "book.epub") }, func(f *fixture) {
			f.kepubifier.KepubifyFn = func(ctx context.Context, inputPath, outputPath string) error {
				return errors.New("broken epub")
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
			err := f.handle(t, tc.payload(t))

			// Assert
			require.NoError(t, err)
			require.Len(t, f.published, 1)
			assert.Equal(t, string(jobs.SubjectFailed), f.published[0].subject)
			assert.NotEmpty(t, f.published[0].msg.ErrorMsg)
			assert.Empty(t, f.uploaded)
			assert.Empty(t, f.deleted)
		})
	}
}

func TestKepubifyUC_Execute_TransientError_ReturnsErrorAndKeepsOriginal(t *testing.T) {
	t.Parallel()

	cloudErr := errors.New("r2 down")
	cases := []struct {
		name  string
		setup func(f *fixture)
	}{
		{"Download", func(f *fixture) {
			f.cloud.DownloadFn = func(ctx context.Context, id, filename string) (io.ReadCloser, error) { return nil, cloudErr }
			f.cloud.SizeFn = func(ctx context.Context, id, filename string) (int64, error) { return 0, cloudErr }
		}},
		{"Upload", func(f *fixture) {
			f.cloud.UploadFn = func(ctx context.Context, id, filename string, body io.Reader) error { return cloudErr }
		}},
		{"Publish", func(f *fixture) {
			f.queue.PublishFn = func(ctx context.Context, subject string, payload any) error { return cloudErr }
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture()
			tc.setup(f)

			// Act
			err := f.handle(t, job(t, "book.epub"))

			// Assert
			assert.ErrorIs(t, err, cloudErr)
			assert.Empty(t, f.deleted)
		})
	}
}

func TestKepubifyUC_Execute_RedeliveredAfterSuccess_Acks(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()
	f.cloud.DownloadFn = func(ctx context.Context, id, filename string) (io.ReadCloser, error) {
		return nil, errors.New("not found")
	}
	f.cloud.SizeFn = func(ctx context.Context, id, filename string) (int64, error) {
		assert.Equal(t, "book.kepub.epub", filename)
		return 10, nil
	}

	// Act
	err := f.handle(t, job(t, "book.epub"))

	// Assert
	require.NoError(t, err)
	assert.Empty(t, f.published)
}

func TestKepubifyUC_Execute_DeleteFails_StillSucceeds(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()
	f.cloud.DeleteFn = func(ctx context.Context, id, filename string) error {
		return errors.New("r2 down")
	}

	// Act
	err := f.handle(t, job(t, "book.epub"))

	// Assert
	require.NoError(t, err)
	require.Len(t, f.published, 1)
	assert.Equal(t, string(jobs.SubjectDone), f.published[0].subject)
}

func TestKepubifyUC_Execute_TransientErrorOnLastAttempt_PublishesFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture()
	f.lastAttempt = true
	f.cloud.UploadFn = func(ctx context.Context, id, filename string, body io.Reader) error {
		return errors.New("r2 down")
	}

	// Act
	err := f.handle(t, job(t, "book.epub"))

	// Assert
	require.NoError(t, err)
	require.Len(t, f.published, 1)
	failed := f.published[0]
	assert.Equal(t, string(jobs.SubjectFailed), failed.subject)
	assert.Equal(t, "result-1", failed.msg.Id)
	assert.Equal(t, "book.epub", failed.msg.Filename)
	assert.NotNil(t, failed.msg.Result)
	assert.Contains(t, failed.msg.ErrorMsg, "r2 down")
	assert.Empty(t, f.deleted)
}
