package libgen_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/infra/libgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMirror answers pings through a real server and searches through SearchFn.
type fakeMirror struct {
	url        string
	SearchFn   func() (models.SearchPage, error)
	DownloadFn func() (*models.BookDownload, error)
	searches   atomic.Int32
	downloads  atomic.Int32
}

func (f *fakeMirror) Search(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
	f.searches.Add(1)
	if f.SearchFn != nil {
		return f.SearchFn()
	}
	return models.SearchPage{Books: []models.Book{{Title: f.url}}}, nil
}

func (f *fakeMirror) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
	f.downloads.Add(1)
	if f.DownloadFn != nil {
		return f.DownloadFn()
	}
	return nil, models.ErrBookNotFound
}

func (f *fakeMirror) URL() string { return f.url }

type fakeMirrorList struct {
	mirrors []libgen.IMirror
	calls   atomic.Int32
}

func (f *fakeMirrorList) Mirrors(ctx context.Context) []libgen.IMirror {
	f.calls.Add(1)
	return append([]libgen.IMirror(nil), f.mirrors...)
}

func newMirror(t *testing.T, delay time.Duration, status int) *fakeMirror {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	return &fakeMirror{url: server.URL}
}

func TestLibgenLibrary_Search_NoMirrorYet_UsesFastest(t *testing.T) {
	t.Parallel()

	// Arrange
	broken := newMirror(t, 0, http.StatusInternalServerError)
	slow := newMirror(t, 200*time.Millisecond, http.StatusOK)
	fast := newMirror(t, 0, http.StatusOK)
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{broken, slow, fast}}, time.Hour)

	// Act
	books, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{{Title: fast.url}}, books.Books)
	assert.Zero(t, broken.searches.Load())
	assert.Zero(t, slow.searches.Load())
}

func TestLibgenLibrary_Search_NoMirrorAnswers_ReturnsErrNoMirror(t *testing.T) {
	t.Parallel()

	// Arrange
	list := &fakeMirrorList{mirrors: []libgen.IMirror{newMirror(t, 0, http.StatusBadGateway)}}
	library := libgen.NewLibgenLibrary(list, time.Hour)

	// Act
	_, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)
	_, errAgain := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	assert.ErrorIs(t, err, models.ErrNoMirror)
	assert.ErrorIs(t, errAgain, models.ErrNoMirror)
	assert.Equal(t, int32(1), list.calls.Load(), "the cooldown avoids a second discovery")
}

func TestLibgenLibrary_Search_MirrorFails_RetriesOnAnotherMirror(t *testing.T) {
	t.Parallel()

	// Arrange
	failing := newMirror(t, 0, http.StatusOK)
	failing.SearchFn = func() (models.SearchPage, error) { return models.SearchPage{}, errors.New("captcha") }
	backup := newMirror(t, 150*time.Millisecond, http.StatusOK)
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{failing, backup}}, 0)

	// Act
	books, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)
	again, errAgain := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{{Title: backup.url}}, books.Books)
	require.NoError(t, errAgain)
	assert.Equal(t, books, again)
	assert.Equal(t, int32(1), failing.searches.Load(), "the backup stays as the current mirror")
}

func TestLibgenLibrary_Search_FailsWithinCooldown_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	searchErr := errors.New("captcha")
	failing := newMirror(t, 0, http.StatusOK)
	failing.SearchFn = func() (models.SearchPage, error) { return models.SearchPage{}, searchErr }
	backup := newMirror(t, 150*time.Millisecond, http.StatusOK)
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{failing, backup}}, time.Hour)

	// Act
	_, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	assert.ErrorIs(t, err, searchErr)
	assert.Zero(t, backup.searches.Load())
}

func TestLibgenLibrary_Search_Concurrent_DiscoversOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	list := &fakeMirrorList{mirrors: []libgen.IMirror{newMirror(t, 100*time.Millisecond, http.StatusOK)}}
	library := libgen.NewLibgenLibrary(list, 0)

	// Act
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			_, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)
			assert.NoError(t, err)
		})
	}
	wg.Wait()

	// Assert
	assert.Equal(t, int32(1), list.calls.Load())
}

func TestLibgenLibrary_Start_Interval_RenewsMirror(t *testing.T) {
	t.Parallel()

	// Arrange
	list := &fakeMirrorList{mirrors: []libgen.IMirror{newMirror(t, 0, http.StatusOK)}}
	library := libgen.NewLibgenLibrary(list, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	library.Start(ctx, 20*time.Millisecond)

	// Assert
	assert.Eventually(t, func() bool { return list.calls.Load() >= 3 }, 2*time.Second, 10*time.Millisecond)
	cancel()
	_, err := library.Search(context.Background(), "dune", models.SearchFieldTitle, 1)
	assert.NoError(t, err)
}

func TestLibgenLibrary_Download_CurrentMirrorFails_TriesTheOthers(t *testing.T) {
	t.Parallel()

	// Arrange
	failing := newMirror(t, 0, http.StatusOK)
	failing.DownloadFn = func() (*models.BookDownload, error) { return nil, errors.New("GET get.php: HTTP 500") }
	missing := newMirror(t, 100*time.Millisecond, http.StatusOK)
	backup := newMirror(t, 100*time.Millisecond, http.StatusOK)
	backup.DownloadFn = func() (*models.BookDownload, error) {
		return &models.BookDownload{Title: "La novia gitana", Ext: "epub"}, nil
	}
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{failing, missing, backup}}, time.Hour)

	// Act
	book, err := library.Download(context.Background(), "0123456789abcdef0123456789abcdef")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "La novia gitana", book.Title)
	assert.Equal(t, int32(1), failing.downloads.Load())
	assert.Equal(t, int32(1), backup.downloads.Load())
}

func TestLibgenLibrary_Download_EveryMirrorFails_ReportsAll(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		firstErr error
		notFound bool
	}{
		{"No mirror has it", models.ErrBookNotFound, true},
		{"One mirror broken", errors.New("GET get.php: HTTP 500"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			first := newMirror(t, 0, http.StatusOK)
			first.DownloadFn = func() (*models.BookDownload, error) { return nil, tc.firstErr }
			second := newMirror(t, 100*time.Millisecond, http.StatusOK)
			library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{first, second}}, time.Hour)

			// Act
			_, err := library.Download(context.Background(), "0123456789abcdef0123456789abcdef")

			// Assert
			require.Error(t, err)
			assert.Equal(t, tc.notFound, errors.Is(err, models.ErrBookNotFound))
			assert.Contains(t, err.Error(), tc.firstErr.Error())
			assert.Equal(t, int32(1), second.downloads.Load())
		})
	}
}
