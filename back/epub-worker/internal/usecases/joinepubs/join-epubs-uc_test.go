package joinepubs_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxBlockSize = 100

type publishedEvent struct {
	subject string
	msg     jobs.Msg
}

type fixture struct {
	lastAttempt bool
	queue       *mocks.MockQueue
	store       *mocks.MockJoinStore
	cloud       *mocks.MockCloudStorage
	merger      *mocks.MockEpubMerger
	published   []publishedEvent
	cleared     bool
	released    bool
}

// newFixture returns mocks where the store always completes the group with children
// and every child weighs 40 bytes in cloud storage.
func newFixture(children []jobs.Msg) *fixture {
	f := &fixture{
		cloud:  &mocks.MockCloudStorage{SizeFn: func(ctx context.Context, id, filename string) (int64, error) { return 40, nil }},
		merger: &mocks.MockEpubMerger{},
	}
	f.queue = &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload any) error {
			f.published = append(f.published, publishedEvent{subject, payload.(jobs.Msg)})
			return nil
		},
	}
	f.store = &mocks.MockJoinStore{
		AddFn: func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
			return children, nil
		},
		ClearFn: func(ctx context.Context, parentId string) error {
			f.cleared = true
			return nil
		},
		ReleaseFn: func(ctx context.Context, parentId string) error {
			f.released = true
			return nil
		},
	}
	return f
}

// handle runs the use case the way main wires it: through jobrun.Handle.
func (f *fixture) handle(t *testing.T, payload []byte) error {
	t.Helper()

	uc := joinepubs.NewJoinEpubsUC(f.queue, f.store, f.cloud, f.merger, maxBlockSize)
	delivery := jobrun.Delivery{Subject: string(jobs.SubjectJoin), Payload: payload, LastAttempt: f.lastAttempt}
	return jobrun.Handle(f.queue, uc.Execute)(context.Background(), delivery)
}

func (f *fixture) bySubject(subject jobs.Subject) []jobs.Msg {
	var msgs []jobs.Msg
	for _, p := range f.published {
		if p.subject == string(subject) {
			msgs = append(msgs, p.msg)
		}
	}
	return msgs
}

func child(id string, index int, kepubify bool) jobs.Msg {
	return jobs.Msg{
		Id:       id,
		UserId:   1,
		Title:    "Ch. " + id,
		Filename: id + ".epub",
		Kepubify: kepubify,
		Parent:   &jobs.Parent{Id: "parent-1", Title: "Manga", Rtl: true, Total: 3, Index: index},
	}
}

func payloadOf(t *testing.T, msg jobs.Msg) []byte {
	t.Helper()

	bytes, err := json.Marshal(msg)
	require.NoError(t, err)
	return bytes
}

func TestJoinEpubsUC_Execute_InvalidJob_PublishesFailed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload func(t *testing.T) []byte
	}{
		{"Invalid json", func(t *testing.T) []byte { return []byte("{") }},
		{"Without parent", func(t *testing.T) []byte { return payloadOf(t, jobs.Msg{Id: "child-1"}) }},
		{"Without total", func(t *testing.T) []byte {
			return payloadOf(t, jobs.Msg{Id: "child-1", Parent: &jobs.Parent{Id: "parent-1"}})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(nil)
			addCalled := false
			f.store.AddFn = func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
				addCalled = true
				return nil, nil
			}

			// Act
			err := f.handle(t, tc.payload(t))

			// Assert
			require.NoError(t, err)
			assert.False(t, addCalled)
			failed := f.bySubject(jobs.SubjectFailed)
			require.Len(t, failed, 1)
			assert.NotEmpty(t, failed[0].ErrorMsg)
		})
	}
}

func TestJoinEpubsUC_Execute_IncompleteGroup_DoesNothing(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(nil)

	// Act
	err := f.handle(t, payloadOf(t, child("1", 0, false)))

	// Assert
	require.NoError(t, err)
	assert.Empty(t, f.published)
	assert.False(t, f.cleared)
}

func TestJoinEpubsUC_Execute_CompleteGroup_PublishesResultsInOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		kepubify        bool
		expectedSubject jobs.Subject
	}{
		{"Done", false, jobs.SubjectDone},
		{"Kepubify", true, jobs.SubjectKepub},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			children := []jobs.Msg{child("3", 2, tc.kepubify), child("1", 0, tc.kepubify), child("2", 1, tc.kepubify)}
			f := newFixture(children)
			var merged []models.EpubMeta
			var mergedTitles [][]string
			f.cloud.DownloadFn = func(ctx context.Context, id, filename string) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(id)), nil
			}
			f.merger = &mocks.MockEpubMerger{MergeFn: func(outputPath string, meta models.EpubMeta, books []models.EpubBook) error {
				merged = append(merged, meta)
				var titles []string
				for _, b := range books {
					titles = append(titles, b.Title)
				}
				mergedTitles = append(mergedTitles, titles)
				return (&mocks.MockEpubMerger{}).Merge(outputPath, meta, books)
			}}
			uploaded := map[string]string{}
			f.cloud.UploadFn = func(ctx context.Context, id, filename string, body io.Reader) error {
				uploaded[id] = filename
				return nil
			}

			// Act
			err := f.handle(t, payloadOf(t, children[0]))

			// Assert
			require.NoError(t, err)
			assert.True(t, f.cleared)
			assert.False(t, f.released)

			require.Len(t, merged, 2)
			assert.Equal(t, models.EpubMeta{Title: "Manga Ch.[1 - 2]", RTL: true}, merged[0])
			assert.Equal(t, models.EpubMeta{Title: "Manga Ch.3", RTL: true}, merged[1])
			assert.Equal(t, [][]string{{"Ch. 1", "Ch. 2"}, {"Ch. 3"}}, mergedTitles)

			results := f.bySubject(tc.expectedSubject)
			require.Len(t, results, 2)
			for i, r := range results {
				assert.Equal(t, joinepubs.ResultId("parent-1", i), r.Id)
				assert.Equal(t, &jobs.Result{ParentId: "parent-1", Index: i, Total: 2}, r.Result)
				assert.Equal(t, r.Filename, uploaded[r.Id])
			}
			assert.Equal(t, "Manga Ch.[1 - 2].epub", results[0].Filename)

			joined := f.bySubject(jobs.SubjectJoined)
			require.Len(t, joined, 3)
			assert.Equal(t, []string{"1", "2", "3"}, []string{joined[0].Id, joined[1].Id, joined[2].Id})
		})
	}
}

func TestJoinEpubsUC_Execute_MergeFails_PublishesFailedResultAndContinues(t *testing.T) {
	t.Parallel()

	// Arrange
	children := []jobs.Msg{child("1", 0, false), child("2", 1, false), child("3", 2, false)}
	f := newFixture(children)
	f.merger = &mocks.MockEpubMerger{MergeFn: func(outputPath string, meta models.EpubMeta, books []models.EpubBook) error {
		if meta.Title == "Manga Ch.[1 - 2]" {
			return errors.New("corrupt epub")
		}
		return (&mocks.MockEpubMerger{}).Merge(outputPath, meta, books)
	}}

	// Act
	err := f.handle(t, payloadOf(t, children[0]))

	// Assert
	require.NoError(t, err)
	assert.True(t, f.cleared)
	failed := f.bySubject(jobs.SubjectFailed)
	require.Len(t, failed, 1)
	assert.Equal(t, joinepubs.ResultId("parent-1", 0), failed[0].Id)
	assert.NotNil(t, failed[0].Result)
	assert.Contains(t, failed[0].ErrorMsg, "corrupt epub")
	assert.Len(t, f.bySubject(jobs.SubjectDone), 1)
	assert.Len(t, f.bySubject(jobs.SubjectJoined), 3)
}

func TestJoinEpubsUC_Execute_TransientError_ReleasesAndReturnsError(t *testing.T) {
	t.Parallel()

	uploadErr := errors.New("r2 down")
	cases := []struct {
		name  string
		setup func(f *fixture)
	}{
		{"Size", func(f *fixture) {
			f.cloud.SizeFn = func(ctx context.Context, id, filename string) (int64, error) { return 0, uploadErr }
		}},
		{"Download", func(f *fixture) {
			f.cloud.DownloadFn = func(ctx context.Context, id, filename string) (io.ReadCloser, error) { return nil, uploadErr }
		}},
		{"Upload", func(f *fixture) {
			f.cloud.UploadFn = func(ctx context.Context, id, filename string, body io.Reader) error { return uploadErr }
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			children := []jobs.Msg{child("1", 0, false)}
			f := newFixture(children)
			tc.setup(f)

			// Act
			err := f.handle(t, payloadOf(t, children[0]))

			// Assert
			assert.ErrorIs(t, err, uploadErr)
			assert.True(t, f.released)
			assert.False(t, f.cleared)
			assert.Empty(t, f.published)
		})
	}
}

func TestJoinEpubsUC_Execute_StoreError_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	storeErr := errors.New("store error")
	f := newFixture(nil)
	f.store.AddFn = func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
		return nil, storeErr
	}

	// Act
	err := f.handle(t, payloadOf(t, child("1", 0, false)))

	// Assert
	assert.ErrorIs(t, err, storeErr)
}

func TestJoinEpubsUC_Execute_CompleteGroup_DeletesMergedSources(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		failedBlock string
		expected    []string
	}{
		{"All merged", "", []string{"1/1.epub", "2/2.epub", "3/3.epub"}},
		{"First block fails", "Manga Ch.[1 - 2]", []string{"3/3.epub"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			children := []jobs.Msg{child("1", 0, false), child("2", 1, false), child("3", 2, false)}
			f := newFixture(children)
			f.merger = &mocks.MockEpubMerger{MergeFn: func(outputPath string, meta models.EpubMeta, books []models.EpubBook) error {
				if meta.Title == tc.failedBlock {
					return errors.New("corrupt epub")
				}
				return (&mocks.MockEpubMerger{}).Merge(outputPath, meta, books)
			}}
			var deleted []string
			f.cloud.DeleteFn = func(ctx context.Context, id, filename string) error {
				deleted = append(deleted, id+"/"+filename)
				return nil
			}

			// Act
			err := f.handle(t, payloadOf(t, children[0]))

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expected, deleted)
		})
	}
}

func TestJoinEpubsUC_Execute_DeleteFails_StillSucceeds(t *testing.T) {
	t.Parallel()

	// Arrange
	children := []jobs.Msg{child("1", 0, false)}
	f := newFixture(children)
	f.cloud.DeleteFn = func(ctx context.Context, id, filename string) error {
		return errors.New("r2 down")
	}

	// Act
	err := f.handle(t, payloadOf(t, children[0]))

	// Assert
	require.NoError(t, err)
	assert.True(t, f.cleared)
	assert.False(t, f.released)
}

func TestJoinEpubsUC_Execute_TransientErrorOnLastAttempt_FailsParentAndClears(t *testing.T) {
	t.Parallel()

	// Arrange
	children := []jobs.Msg{child("1", 0, false)}
	f := newFixture(children)
	f.lastAttempt = true
	f.cloud.SizeFn = func(ctx context.Context, id, filename string) (int64, error) {
		return 0, errors.New("r2 down")
	}

	// Act
	err := f.handle(t, payloadOf(t, children[0]))

	// Assert
	require.NoError(t, err)
	assert.True(t, f.cleared)
	assert.False(t, f.released)
	failed := f.bySubject(jobs.SubjectFailed)
	require.Len(t, failed, 1)
	assert.Equal(t, "parent-1", failed[0].Id)
	assert.Contains(t, failed[0].ErrorMsg, "r2 down")
}

func TestJoinEpubsUC_Execute_InvalidChildren_FailsThemAndMergesTheRest(t *testing.T) {
	t.Parallel()

	// Arrange
	kepub := child("2", 1, false)
	kepub.Filename = "2.kepub.epub"
	children := []jobs.Msg{child("1", 0, false), kepub, child("3", 2, false)}
	f := newFixture(children)

	// Act
	err := f.handle(t, payloadOf(t, children[0]))

	// Assert
	require.NoError(t, err)
	failed := f.bySubject(jobs.SubjectFailed)
	require.Len(t, failed, 1)
	assert.Equal(t, "2", failed[0].Id)
	assert.Contains(t, failed[0].ErrorMsg, "kepub")

	done := f.bySubject(jobs.SubjectDone)
	require.Len(t, done, 1)
	assert.Equal(t, "Manga Ch.[1 - 3]", done[0].Title)
	assert.Len(t, f.bySubject(jobs.SubjectJoined), 2)
}

func TestJoinEpubsUC_Execute_NoValidChildren_FailsParent(t *testing.T) {
	t.Parallel()

	// Arrange
	invalid := child("1", 0, false)
	invalid.Filename = "1.pdf"
	f := newFixture([]jobs.Msg{invalid})

	// Act
	err := f.handle(t, payloadOf(t, invalid))

	// Assert
	require.NoError(t, err)
	assert.True(t, f.cleared)
	failed := f.bySubject(jobs.SubjectFailed)
	require.Len(t, failed, 2)
	assert.Equal(t, "1", failed[0].Id)
	assert.Equal(t, "parent-1", failed[1].Id)
}
