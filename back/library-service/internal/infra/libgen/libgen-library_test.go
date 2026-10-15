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

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/libgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMirror answers pings through a real server and searches through SearchFn.
type fakeMirror struct {
	url      string
	SearchFn func() ([]models.Book, error)
	searches atomic.Int32
}

func (f *fakeMirror) Search(ctx context.Context, query string) ([]models.Book, error) {
	f.searches.Add(1)
	if f.SearchFn != nil {
		return f.SearchFn()
	}
	return []models.Book{{Title: f.url}}, nil
}

func (f *fakeMirror) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
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
	books, err := library.Search(context.Background(), "dune")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{{Title: fast.url}}, books)
	assert.Zero(t, broken.searches.Load())
	assert.Zero(t, slow.searches.Load())
}

func TestLibgenLibrary_Search_NoMirrorAnswers_ReturnsErrNoMirror(t *testing.T) {
	t.Parallel()

	// Arrange
	list := &fakeMirrorList{mirrors: []libgen.IMirror{newMirror(t, 0, http.StatusBadGateway)}}
	library := libgen.NewLibgenLibrary(list, time.Hour)

	// Act
	_, err := library.Search(context.Background(), "dune")
	_, errAgain := library.Search(context.Background(), "dune")

	// Assert
	assert.ErrorIs(t, err, models.ErrNoMirror)
	assert.ErrorIs(t, errAgain, models.ErrNoMirror)
	assert.Equal(t, int32(1), list.calls.Load(), "the cooldown avoids a second discovery")
}

func TestLibgenLibrary_Search_MirrorFails_RetriesOnAnotherMirror(t *testing.T) {
	t.Parallel()

	// Arrange
	failing := newMirror(t, 0, http.StatusOK)
	failing.SearchFn = func() ([]models.Book, error) { return nil, errors.New("captcha") }
	backup := newMirror(t, 150*time.Millisecond, http.StatusOK)
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{failing, backup}}, 0)

	// Act
	books, err := library.Search(context.Background(), "dune")
	again, errAgain := library.Search(context.Background(), "dune")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{{Title: backup.url}}, books)
	require.NoError(t, errAgain)
	assert.Equal(t, books, again)
	assert.Equal(t, int32(1), failing.searches.Load(), "the backup stays as the current mirror")
}

func TestLibgenLibrary_Search_FailsWithinCooldown_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	searchErr := errors.New("captcha")
	failing := newMirror(t, 0, http.StatusOK)
	failing.SearchFn = func() ([]models.Book, error) { return nil, searchErr }
	backup := newMirror(t, 150*time.Millisecond, http.StatusOK)
	library := libgen.NewLibgenLibrary(&fakeMirrorList{mirrors: []libgen.IMirror{failing, backup}}, time.Hour)

	// Act
	_, err := library.Search(context.Background(), "dune")

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
			_, err := library.Search(context.Background(), "dune")
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
	_, err := library.Search(context.Background(), "dune")
	assert.NoError(t, err)
}
